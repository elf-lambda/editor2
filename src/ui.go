package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type UIState struct {
	ActiveMenu     string
	ModalOpen      string
	InputBoxes     []*InputBox
	CurrentView    string
	ShowNotesPanel bool
	Notes          []string
	NotesScroll    int
	NotesPath      string
	IsFolderView   bool

	ShowFilePicker   bool
	FileEntries      []FileEntry
	CurrentPath      string
	SelectedFile     string
	SelectedIsFolder bool
}

type InputBox struct {
	Rect     rl.Rectangle
	Text     string
	Focused  bool
	MaxChars int
}

// new "modern" colors - basically the theme
var (
	ModernDarkBg     = rl.NewColor(30, 32, 60, 255)    // Dark background
	ModernDarkButton = rl.NewColor(30, 32, 48, 150)    // Dark button
	ModernDark       = rl.NewColor(30, 32, 48, 255)    // Dark background
	ModernMedium     = rl.NewColor(52, 58, 84, 255)    // Medium background
	ModernLight      = rl.NewColor(73, 82, 122, 255)   // Light accent
	ModernAccent     = rl.NewColor(108, 117, 255, 255) // Primary accent
	ModernSuccess    = rl.NewColor(72, 187, 120, 255)  // Success green
	ModernDanger     = rl.NewColor(245, 101, 101, 255) // Danger red
	ModernText       = rl.NewColor(226, 232, 240, 255) // Light text
	ModernTextDim    = rl.NewColor(160, 174, 192, 255) // Dimmed text
	ModernShadow     = rl.NewColor(0, 0, 0, 50)        // Subtle shadow
)

func drawShadow(x, y, width, height, offset, blur float32) {
	// just draw multiple rectangles with lowered alpha to create blur effect
	for i := 0; i < int(blur); i++ {
		alpha := uint8(float32(ModernShadow.A) * (1.0 - float32(i)/blur))
		shadowColor := rl.NewColor(0, 0, 0, alpha)
		rl.DrawRectangle(int32(x+offset+float32(i)), int32(y+offset+float32(i)), int32(width), int32(height), shadowColor)
	}
}

func (box *InputBox) Draw() {
	// draw shadow
	drawShadow(box.Rect.X, box.Rect.Y, box.Rect.Width, box.Rect.Height, 2, 3)

	bg := ModernMedium
	if box.Focused {
		bg = ModernLight
	}

	// draw panel
	rl.DrawRectangle(int32(box.Rect.X), int32(box.Rect.Y), int32(box.Rect.Width), int32(box.Rect.Height), bg)

	// draw border
	borderColor := ModernLight
	if box.Focused {
		borderColor = ModernAccent
	}
	rl.DrawRectangleLines(int32(box.Rect.X), int32(box.Rect.Y), int32(box.Rect.Width), int32(box.Rect.Height), borderColor)

	// draw text with better positioning
	DrawText(box.Text, int(box.Rect.X)+12, int(box.Rect.Y)+8, 10, rl.DrawPixel, "white")
}

func (box *InputBox) HandleInput() {
	// autofocus
	box.Focused = true

	if box.Focused {
		char := rl.GetCharPressed()
		for char > 0 {
			if char >= 32 && char <= 126 && len(box.Text) < box.MaxChars {
				box.Text += string(rune(char))
			}
			char = rl.GetCharPressed()
		}
		if (rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyDown(rl.KeyBackspace)) && len(box.Text) > 0 {
			box.Text = box.Text[:len(box.Text)-1]
			time.Sleep(150 * time.Millisecond)
			return
		}
	}
}

func DrawMenuBar(ui *UIState) {
	menuHeight := editorTopPadding

	// Modern gradient-like background
	rl.DrawRectangle(0, 0, int32(windowWidth), int32(menuHeight), ModernDark)
	rl.DrawRectangle(0, int32(menuHeight-2), int32(windowWidth), 2, ModernAccent)

	if DrawModernButton("File", 10, 6, 60, int32(menuHeight-12), ModernText, ModernAccent, ModernLight, ModernDark, true) {
		if ui.ActiveMenu == "File" {
			ui.ActiveMenu = ""
		} else {
			ui.ActiveMenu = "File"
		}
	}
	if DrawModernButton("Notes", 80, 6, 70, int32(menuHeight-12), ModernText, ModernAccent, ModernLight, ModernDark, true) {
		if ui.ShowNotesPanel {
			ui.ShowNotesPanel = false
		} else {
			ui.NotesPath = ""
			ui.Notes = listNoteFiles("notes", true) // list only folders
			ui.IsFolderView = true
			ui.ShowNotesPanel = true
			ui.NotesScroll = 0
		}
	}

	if DrawModernButton("Create Note", 160, 6, 110, int32(menuHeight-12), ModernText, ModernSuccess, ModernLight, ModernDark, true) {
		ui.ModalOpen = "CreateNote"
		ui.InputBoxes = []*InputBox{
			{
				Rect:     rl.NewRectangle(150, 150, 300, 40),
				Text:     "",
				MaxChars: 64,
			},
		}
	}
	if DrawModernButton("Delete Note", 280, 6, 110, int32(menuHeight-12), ModernText, ModernDanger, ModernLight, ModernDark, true) {
		if strings.Contains(currentFile, "notes/") {
			deleteFile(currentFile)
			clearTextGrid()
		} else {
			editorStatus = "Not a Note"
		}
	}

	if ui.ActiveMenu == "File" {
		DrawDropdown("File", 10, int32(menuHeight), ui)
	}
}

func DrawNotesPanel(ui *UIState) {
	panelX, panelY := int32(windowWidth/2/3), int32(windowHeight/2/3)
	panelW, panelH := int32(400), int32(300)
	entryHeight := int32(28)
	maxVisible := int(panelH-80) / int(entryHeight)

	// mouse wheel scrolling
	mouseWheel := rl.GetMouseWheelMove()
	if mouseWheel != 0 {
		// check if mouse is over the panel
		mouseX := rl.GetMouseX()
		mouseY := rl.GetMouseY()
		if mouseX >= panelX && mouseX <= panelX+panelW && mouseY >= panelY && mouseY <= panelY+panelH {
			ui.NotesScroll -= int(mouseWheel)

			// clamp scroll bounds
			maxScroll := len(ui.Notes) - maxVisible
			if maxScroll < 0 {
				maxScroll = 0
			}
			if ui.NotesScroll < 0 {
				ui.NotesScroll = 0
			}
			if ui.NotesScroll > maxScroll {
				ui.NotesScroll = maxScroll
			}
		}
	}

	// draw shadow
	drawShadow(float32(panelX), float32(panelY), float32(panelW), float32(panelH), 3, 4)

	// draw panel
	rl.DrawRectangle(panelX, panelY, panelW, panelH, ModernMedium)

	// draw border around the entire panel
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, rl.DarkGray)

	// header section
	rl.DrawRectangle(panelX, panelY, panelW, 50, ModernDark)

	DrawText("Notes", int(panelX)+16, int(panelY)+16, 12, rl.DrawPixel, "white")

	// back Button if not root ""
	if ui.NotesPath != "" {
		if DrawModernButton("Back", panelX+16, panelY+panelH-30, 70, 24, ModernText, ModernAccent, ModernLight, ModernDarkButton, true) {
			ui.NotesPath = ""
			ui.Notes = listNoteFiles("notes", true)
			ui.IsFolderView = true
			ui.NotesScroll = 0
			return
		}
	}

	if DrawModernButton("Close", panelX+panelW-86, panelY+panelH-30, 70, 24, ModernText, ModernDanger, ModernLight, ModernDarkButton, true) {
		ui.ShowNotesPanel = false
		return
	}

	// draw scroll bar if there are more items than visible
	if len(ui.Notes) > maxVisible {
		scrollBarX := panelX + panelW - 12
		scrollBarY := panelY + 60
		scrollBarH := panelH - 110

		// scroll track
		rl.DrawRectangle(scrollBarX, scrollBarY, 6, scrollBarH, rl.Gray)

		// scroll thumb
		thumbHeight := int32(30)
		maxScroll := len(ui.Notes) - maxVisible
		scrollRatio := float32(ui.NotesScroll) / float32(maxScroll)
		thumbY := scrollBarY + int32(scrollRatio*float32(scrollBarH-thumbHeight))

		rl.DrawRectangle(scrollBarX, thumbY, 6, thumbHeight, rl.White)
	}

	// visible list
	startIdx := ui.NotesScroll
	endIdx := startIdx + maxVisible
	if endIdx > len(ui.Notes) {
		endIdx = len(ui.Notes)
	}

	// fmt.Println("startIDx =", startIdx)
	// fmt.Println("endidx =", endIdx)
	// fmt.Println("len ui.notes", len(ui.Notes))
	for i := startIdx; i < endIdx; i++ {
		y := panelY + 60 + int32((i-startIdx)*int(entryHeight))
		var entry string

		if len(ui.Notes) == 0 {
			continue
		} else if i > len(ui.Notes) {
			continue
		} else if len(ui.Notes) < endIdx {
			entry = ui.Notes[i-1]
		} else {
			entry = ui.Notes[i]
		}
		// fmt.Println("len ui.notes", len(ui.Notes))

		// clamp the file name > long_file_name.txt -> long_file_n... for example
		entryName := clampName(entry, panelW-40-(CHAR_IMAGE_WIDTH*4), CHAR_IMAGE_WIDTH)
		if DrawModernButton(entryName, panelX+16, y, panelW-32, 24, ModernText, ModernAccent, ModernLight, ModernMedium, false) {
			if ui.IsFolderView {
				// folder clicked
				ui.NotesPath = entry
				ui.Notes = listNoteFiles("notes/"+entry, false)
				ui.IsFolderView = false
				ui.NotesScroll = 0
			} else {
				// file clicked
				fullPath := "/home/void/notes/" + ui.NotesPath + entry
				clearTextGrid()
				loadFileIntoTextGrid(fullPath)
				cursor.reset()
				currentFile = fullPath
				editorStatus = "Loaded Note: " + entry
				ui.ShowNotesPanel = false
			}
		}
	}
	ensureCursorVisible(cursor)
}

func DrawFilePickerPanel(ui *UIState) {
	panelX, panelY := int32(windowWidth/2/3), int32(windowHeight/2/3)
	panelW, panelH := int32(400), int32(300)
	entryHeight := int32(28)
	maxVisible := int(panelH-80) / int(entryHeight)

	// mouse wheel scrolling
	mouseWheel := rl.GetMouseWheelMove()
	if mouseWheel != 0 {
		mouseX := rl.GetMouseX()
		mouseY := rl.GetMouseY()
		if mouseX >= panelX && mouseX <= panelX+panelW && mouseY >= panelY && mouseY <= panelY+panelH {
			ui.NotesScroll -= int(mouseWheel)

			maxScroll := len(ui.FileEntries) - maxVisible
			if maxScroll < 0 {
				maxScroll = 0
			}
			if ui.NotesScroll < 0 {
				ui.NotesScroll = 0
			}
			if ui.NotesScroll > maxScroll {
				ui.NotesScroll = maxScroll
			}
		}
	}

	// draw shadow
	drawShadow(float32(panelX), float32(panelY), float32(panelW), float32(panelH), 4, 8)

	// draw panel
	rl.DrawRectangle(panelX, panelY, panelW, panelH, ModernMedium)
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, rl.DarkGray)

	// header section
	rl.DrawRectangle(panelX, panelY, panelW, 50, ModernDark)
	DrawText("File Picker", int(panelX)+16, int(panelY)+16, 12, rl.DrawPixel, "white")

	// show current path
	if ui.CurrentPath == "" {
		ui.CurrentPath, _ = filepath.Abs(".") // Start here TODO: rework
	}

	DrawText("Path: "+ui.CurrentPath, int(panelX)+16, int(panelY)+35, 8, rl.DrawPixel, "white")

	// back button (if not at root)
	if ui.CurrentPath != "/" {
		if DrawModernButton("Back", panelX+16, panelY+panelH-30, 70, 24, ModernText, ModernAccent, ModernLight, ModernDark, true) {
			// Go up one directory
			if strings.Contains(ui.CurrentPath, "/") {
				ui.CurrentPath = filepath.Dir(ui.CurrentPath)
				if ui.CurrentPath == "." {
					ui.CurrentPath = ""
				}
			} else {
				ui.CurrentPath = ""
			}
			ui.FileEntries = listAllFiles(ui.CurrentPath)
			ui.NotesScroll = 0
			return
		}
	}

	// close button
	if DrawModernButton("Close", panelX+panelW-86, panelY+panelH-30, 70, 24, ModernText, ModernDanger, ModernLight, ModernDark, true) {
		ui.ShowFilePicker = false
		return
	}

	// select button (only show if we have a selected file)
	if ui.SelectedFile != "" && !ui.SelectedIsFolder {
		if DrawModernButton("Select", panelX+panelW-166, panelY+panelH-30, 70, 24, ModernText, ModernSuccess, ModernLight, ModernDark, true) {
			// handle file selection here
			fullPath := filepath.Join(ui.CurrentPath, ui.SelectedFile)
			// clearTextGrid()
			code, err := loadFileIntoTextGrid(fullPath)
			if code == -1 || err != nil {
				currentFile = "Untitled"
				loadStringIntoTextGrid(err.Error())
				ensureCursorVisible(cursor)
				ui.ShowFilePicker = false
				return
			}
			// cursor.reset()
			currentFile = fullPath
			editorStatus = "Loaded: " + ui.SelectedFile
			ui.ShowFilePicker = false
			ensureCursorVisible(cursor)
			return
		}
	}

	// scroll bar
	if len(ui.FileEntries) > maxVisible {
		scrollBarX := panelX + panelW - 12
		scrollBarY := panelY + 60
		scrollBarH := panelH - 110

		rl.DrawRectangle(scrollBarX, scrollBarY, 6, scrollBarH, rl.Gray)

		thumbHeight := int32(30)
		maxScroll := len(ui.FileEntries) - maxVisible
		scrollRatio := float32(ui.NotesScroll) / float32(maxScroll)
		thumbY := scrollBarY + int32(scrollRatio*float32(scrollBarH-thumbHeight))
		rl.DrawRectangle(scrollBarX, thumbY, 6, thumbHeight, rl.White)
	}

	// file/folder list
	startIdx := ui.NotesScroll
	endIdx := startIdx + maxVisible
	if endIdx > len(ui.FileEntries) {
		endIdx = len(ui.FileEntries)
	}

	for i := startIdx; i < endIdx; i++ {
		if i >= len(ui.FileEntries) {
			continue
		}

		entry := ui.FileEntries[i]
		y := panelY + 60 + int32((i-startIdx)*int(entryHeight))

		// different styling for folders vs files
		var displayName string
		var textColor rl.Color = ModernText

		if entry.IsFolder {
			displayName = "[dir] " + entry.Name
			textColor = ModernAccent
		} else {
			displayName = "[file] " + entry.Name
		}

		// highlight selected item
		bgColor := ModernMedium
		if ui.SelectedFile == entry.Name {
			bgColor = ModernLight
		}

		// draw all elements (files/folders)
		// entryName := clampName(displayName, panelW-50-(CHAR_IMAGE_WIDTH*4), CHAR_IMAGE_WIDTH)
		entryName := clampName(displayName, panelW-40-(CHAR_IMAGE_WIDTH*4), CHAR_IMAGE_WIDTH)
		// if DrawModernButton(entryName, panelX+16, y, panelW-32, 24, ModernText, ModernAccent, ModernLight, ModernMedium, false) {

		if DrawModernButton(entryName, panelX+16, y, panelW-32, 24, textColor, ModernAccent, ModernLight, bgColor, false) {
			ui.SelectedFile = entry.Name
			ui.SelectedIsFolder = entry.IsFolder

			if entry.IsFolder {
				// Double-click or enter folder
				ui.CurrentPath = filepath.Join(ui.CurrentPath, entry.Name)
				ui.FileEntries = listAllFiles(ui.CurrentPath)
				ui.NotesScroll = 0
				ui.SelectedFile = ""
				ui.SelectedIsFolder = false
			}
		}
	}

}

func DrawDropdown(menu string, x, y int32, ui *UIState) {
	options := []string{"Open...", "Open Pick", "Save", "Save As...", "New"}
	dropdownW := int32(120)
	dropdownH := int32(len(options) * 32)

	// draw shadow
	drawShadow(float32(x), float32(y), float32(dropdownW), float32(dropdownH), 2, 4)

	// draw dropdown panel
	rl.DrawRectangle(x, y, dropdownW, dropdownH, ModernMedium)

	for i, opt := range options {
		btnY := y + int32(i*32)
		if DrawModernButton(opt, x+4, btnY+4, dropdownW-8, 24, ModernText, ModernAccent, ModernLight, ModernMedium, true) {
			switch opt {
			case "Open...":
				ui.ModalOpen = "OpenFile"
				ui.InputBoxes = []*InputBox{
					{
						Rect:     rl.NewRectangle(150, 150, 300, 40),
						Text:     "",
						MaxChars: 64,
					},
				}
			case "Save":
				fmt.Println("Save triggered (implement me!)")
				if currentFile == "Untitled" {
					ui.ModalOpen = "SaveAs"
					ui.InputBoxes = []*InputBox{
						{
							Rect:     rl.NewRectangle(150, 150, 300, 40),
							Text:     "",
							MaxChars: 64,
						},
					}
					continue
				}
				err := saveTextGridToFile(currentFile)
				if err != nil {
					currentFile = "Untitled"
					loadStringIntoTextGrid(err.Error())
					ui.ActiveMenu = ""
					return
				}
				editorStatus = "Saved File!"
			case "Save As...":
				ui.ModalOpen = "SaveAs"
				ui.InputBoxes = []*InputBox{
					{
						Rect:     rl.NewRectangle(150, 150, 300, 40),
						Text:     "",
						MaxChars: 64,
					},
				}
			case "New":
				clearTextGrid()
				cursor.reset()
				// printGrid(cursor)
			case "Open Pick":
				if ui.ShowFilePicker {
					ui.ShowFilePicker = false
				} else {
					ui.ShowFilePicker = true
					ui.CurrentPath = ""
					ui.FileEntries = listAllFiles(".")
					ui.NotesScroll = 0
				}
			}
			ui.ActiveMenu = ""
		}
	}
}

func DrawModal(ui *UIState) {
	if ui.ModalOpen == "OpenFile" {
		modalX := int32(100)
		modalY := int32(50)
		modalW := int32(windowWidth - 200)
		modalH := int32(windowHeight - 150)

		// TODO: Rework this
		// draw modal blur
		rl.DrawRectangle(0, 0, int32(windowWidth), int32(windowHeight), rl.NewColor(0, 0, 0, 128))

		// draw shadow
		drawShadow(float32(modalX), float32(modalY), float32(modalW), float32(modalH), 6, 12)

		// draw modal panel
		rl.DrawRectangle(modalX, modalY, modalW, modalH, ModernMedium)

		// header
		rl.DrawRectangle(modalX, modalY, modalW, 60, ModernDark)

		DrawText("Open File", int(modalX)+20, int(modalY)+20, 14, rl.DrawPixel, "white")

		for _, ib := range ui.InputBoxes {
			ib.Draw()
			ib.HandleInput()
		}

		if DrawModernButton("Cancel", modalX+modalW-180, modalY+modalH-50, 80, 32, ModernText, ModernDanger, ModernLight, ModernMedium, true) {
			ui.ModalOpen = ""
		}

		if DrawModernButton("Open", modalX+modalW-90, modalY+modalH-50, 80, 32, ModernText, ModernSuccess, ModernLight, ModernMedium, true) {
			if len(ui.InputBoxes) > 0 {
				filename := ui.InputBoxes[0].Text
				fmt.Printf("open: %s\n", filename)
				res, err := loadFileIntoTextGrid(filename)
				if err != nil {
					currentFile = "Untitled"
					// clearTextGrid()
					// cursor.reset()
					loadStringIntoTextGrid(err.Error())
					ui.ModalOpen = ""
					return
				}
				fmt.Println("Loaded ", res, " bytes into text grid")
			}
			ui.ModalOpen = ""
		}
	}
	if ui.ModalOpen == "SaveAs" {
		modalX := int32(100)
		modalY := int32(50)
		modalW := int32(windowWidth - 200)
		modalH := int32(windowHeight - 150)

		// draw modal blur
		rl.DrawRectangle(0, 0, int32(windowWidth), int32(windowHeight), rl.NewColor(0, 0, 0, 128))

		// draw shadow
		drawShadow(float32(modalX), float32(modalY), float32(modalW), float32(modalH), 6, 12)

		// draw modal panel
		rl.DrawRectangle(modalX, modalY, modalW, modalH, ModernMedium)

		// header
		rl.DrawRectangle(modalX, modalY, modalW, 60, ModernDark)

		DrawText("Save As", int(modalX)+20, int(modalY)+20, 14, rl.DrawPixel, "white")

		for _, ib := range ui.InputBoxes {
			ib.Draw()
			ib.HandleInput()
		}

		if DrawModernButton("Cancel", modalX+modalW-180, modalY+modalH-50, 80, 32, ModernText, ModernDanger, ModernLight, ModernMedium, true) {
			ui.ModalOpen = ""
		}

		if DrawModernButton("Save", modalX+modalW-90, modalY+modalH-50, 80, 32, ModernText, ModernSuccess, ModernLight, ModernMedium, true) {
			if len(ui.InputBoxes) > 0 {
				filename := ui.InputBoxes[0].Text
				err := saveTextGridToFile(filename)
				if err != nil {
					currentFile = "Untitled"
					loadStringIntoTextGrid(err.Error())
					ui.ModalOpen = ""
					return
				} else {
					fmt.Println("File saved as", filename)
					currentFile = filename
				}
			}
			ui.ModalOpen = ""
		}
	}
	if ui.ModalOpen == "CreateNote" {
		modalX := int32(100)
		modalY := int32(50)
		modalW := int32(windowWidth - 200)
		modalH := int32(windowHeight - 150)

		// draw modal blur
		rl.DrawRectangle(0, 0, int32(windowWidth), int32(windowHeight), rl.NewColor(0, 0, 0, 128))

		// draw shadow
		drawShadow(float32(modalX), float32(modalY), float32(modalW), float32(modalH), 6, 12)

		// draw modal panel
		rl.DrawRectangle(modalX, modalY, modalW, modalH, ModernMedium)

		// header
		rl.DrawRectangle(modalX, modalY, modalW, 60, ModernDark)

		DrawText("Create Note", int(modalX)+20, int(modalY)+20, 14, rl.DrawPixel, "white")

		for _, ib := range ui.InputBoxes {
			ib.Draw()
			ib.HandleInput()
		}

		if DrawModernButton("Cancel", modalX+modalW-180, modalY+modalH-50, 80, 32, ModernText, ModernDanger, ModernLight, ModernMedium, true) {
			ui.ModalOpen = ""
		}

		if DrawModernButton("Create", modalX+modalW-90, modalY+modalH-50, 80, 32, ModernText, ModernSuccess, ModernLight, ModernMedium, true) {
			if len(ui.InputBoxes) > 0 {
				filename := ui.InputBoxes[0].Text

				now := time.Now()
				folderName := now.Format("01-2006")
				baseDir := "/home/void/"
				// if err != nil {
				// 	fmt.Printf("Error getting current working directory: %v\n", err)
				// 	return
				// }

				folderPath := filepath.Join(baseDir, "notes/"+folderName)
				fmt.Printf("Attempting to create folder: %s\n", folderPath)
				_, err := os.Stat(folderPath)
				if os.IsNotExist(err) {
					err = os.MkdirAll(folderPath, 0755)
					if err != nil {
						fmt.Printf("Error creating folder %s: %v\n", folderPath, err)
						return
					}
					fmt.Printf("Folder '%s' created successfully.\n", folderPath)
				} else if err != nil {
					fmt.Printf("Error checking folder existence %s: %v\n", folderPath, err)
					return
				} else {
					fmt.Printf("Folder '%s' already exists.\n", folderPath)
				}

				fileName := now.Format("02-01-2006")
				var filePath string
				if filename == "" {
					filePath = filepath.Join(folderPath, fileName)
				} else {
					filePath = filepath.Join(folderPath, fileName+" - "+filename)
				}

				fmt.Printf("Attempting to create file: %s\n", filePath)
				filePath += ".txt"

				_, err = os.Stat(filePath)
				if os.IsNotExist(err) {
					file, err := os.Create(filePath)
					if err != nil {
						fmt.Printf("Error creating file %s: %v\n", filePath, err)
						return
					}
					defer file.Close()

					fmt.Printf("File '%s' created successfully.\n", filePath)

				} else if err != nil {
					fmt.Printf("Error checking file existence %s: %v\n", filePath, err)
					return
				} else {
					fmt.Printf("File '%s' already exists.\n", filePath)
					ui.ModalOpen = ""
					return
				}

				err = saveTextGridToFile(filePath)
				if err != nil {
					currentFile = "Untitled"
					loadStringIntoTextGrid(err.Error())
					ui.ModalOpen = ""
					return
				} else {
					fmt.Println("File saved as", filePath)
					currentFile = filePath
				}
			}
			ui.ModalOpen = ""
		}
	}
}

func DrawStatusBar(cursor Cursor) {
	barHeight := editorBottomPadding

	// modern gradient background
	rl.DrawRectangle(0, int32(windowHeight)-int32(barHeight), int32(windowWidth), int32(barHeight), ModernDark)
	rl.DrawRectangle(0, int32(windowHeight)-int32(barHeight), int32(windowWidth), 2, ModernAccent)

	status := fmt.Sprintf("Ln %d, Col %d Buffer: %s | Status: %s", cursor.y+1, cursor.x+1, currentFile, editorStatus)
	DrawText(status, 12, windowHeight-barHeight+5, CHAR_IMAGE_WIDTH, rl.DrawPixel, "white")
}

func DrawModernButton(label string, x, y, w, h int32, textColor rl.Color, pressColor, hoverColor, idleColor rl.Color, padding bool) bool {
	mouseX := rl.GetMouseX()
	mouseY := rl.GetMouseY()
	mouseOver := mouseX >= x && mouseX <= x+w && mouseY >= y && mouseY <= y+h
	pressed := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	var bgColor rl.Color
	if pressed && mouseOver {
		bgColor = pressColor
	} else if mouseOver {
		bgColor = hoverColor
	} else {
		bgColor = idleColor
	}

	// draw square button
	rl.DrawRectangle(x, y, w, h, bgColor)

	// simple text positioning - always left aligned with padding
	// keeping padding bool as i cant be bothered to rewrite
	// TODO: remove unused padding bool
	DrawText(label, int(x)+8, int(y)+int(h)/2-5, CHAR_IMAGE_WIDTH, rl.DrawPixel, "white")

	return mouseOver && pressed
}

func isCellSelected(x, y int) bool {
	if !selection.Active {
		return false
	}

	startX, startY := selection.StartX, selection.StartY
	endX, endY := selection.EndX, selection.EndY

	// normalize coordinates
	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	if y < startY || y > endY {
		return false
	}

	if y == startY && y == endY {
		return x >= startX && x <= endX
	} else if y == startY {
		return x >= startX
	} else if y == endY {
		return x <= endX
	}
	return true
}
