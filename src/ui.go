package main

import (
	"fmt"
	"os"
	"path/filepath"
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
}

type InputBox struct {
	Rect     rl.Rectangle
	Text     string
	Focused  bool
	MaxChars int
}

func listNoteFiles(dir string, foldersOnly bool) []string {
	os.MkdirAll("notes", os.ModePerm)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var entries []string
	for _, f := range files {
		if foldersOnly && f.IsDir() {
			entries = append(entries, f.Name()+"/")
		} else if !foldersOnly && !f.IsDir() && filepath.Ext(f.Name()) == ".txt" {
			entries = append(entries, f.Name())
		}
	}

	fmt.Println(entries)
	return entries
}

func (box *InputBox) Draw() {
	bg := rl.DarkBrown
	if box.Focused {
		bg = rl.LightGray
	}

	rl.DrawRectangleRec(box.Rect, bg)
	rl.DrawRectangleLines(int32(box.Rect.X), int32(box.Rect.Y), int32(box.Rect.Width), int32(box.Rect.Height), rl.Black)
	DrawText(box.Text, int(box.Rect.X)+5, int(box.Rect.Y)+5, 10, rl.DrawPixel, "black")
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
		if rl.IsKeyPressed(rl.KeyBackspace) && len(box.Text) > 0 {
			box.Text = box.Text[:len(box.Text)-1]
			return
		}
	}
}

func DrawMenuBar(ui *UIState) {
	menuHeight := editorTopPadding
	rl.DrawRectangle(0, 0, int32(windowWidth), int32(menuHeight), rl.DarkBlue)

	if DrawButton("File", 0, 0, 50, int32(menuHeight), "white", rl.DarkGreen, rl.Gray, rl.DarkBlue) {
		if ui.ActiveMenu == "File" {
			ui.ActiveMenu = ""
		} else {
			ui.ActiveMenu = "File"
		}
	}
	if DrawButton("Notes", 50, 0, 50, int32(menuHeight), "white", rl.DarkGreen, rl.Gray, rl.DarkBlue) {
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

	if DrawButton("Create Note", 100, 0, 120, int32(menuHeight), "white", rl.DarkGreen, rl.Gray, rl.DarkBlue) {
		ui.ModalOpen = "CreateNote"
		ui.InputBoxes = []*InputBox{
			{
				Rect:     rl.NewRectangle(150, 150, 300, 30),
				Text:     "",
				MaxChars: 64,
			},
		}
	}
	if DrawButton("Delete Note", 220, 0, 120, int32(menuHeight), "white", rl.DarkGreen, rl.Gray, rl.DarkBlue) {
		deleteFile(currentFile)
		clearTextGrid()
	}

	if ui.ActiveMenu == "File" {
		DrawDropdown("File", 0, int32(menuHeight), ui)
	}
}

func DrawNotesPanel(ui *UIState) {
	panelX, panelY := int32(windowWidth/2/2), int32(windowHeight/2/3)
	panelW, panelH := int32(300), int32(250)
	entryHeight := int32(25)
	maxVisible := int(panelH-60) / int(entryHeight)

	// mouse wheel scrolling
	mouseWheel := rl.GetMouseWheelMove()
	if mouseWheel != 0 {
		// check if mouse is over the panel
		mouseX := rl.GetMouseX()
		mouseY := rl.GetMouseY()
		if mouseX >= panelX && mouseX <= panelX+panelW && mouseY >= panelY && mouseY <= panelY+panelH {
			ui.NotesScroll -= int(mouseWheel)

			// Clamp scroll bounds
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

	rl.DrawRectangle(panelX, panelY, panelW, panelH, rl.DarkGray)
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, rl.Black)
	DrawText("Notes", int(panelX)+10, int(panelY)+10, 10, rl.DrawPixel, "white")

	// back Button if not root ""
	if ui.NotesPath != "" {
		if DrawButton("Back", panelX+10, panelY+panelH-30, 60, 20, "white", rl.DarkBlue, rl.Gray, rl.DarkGray) {
			ui.NotesPath = ""
			ui.Notes = listNoteFiles("notes", true)
			ui.IsFolderView = true
			ui.NotesScroll = 0
			return
		}
	}

	if DrawButton("Close", panelX+panelW-80, panelY+panelH-30, 70, 20, "white", rl.Red, rl.Gray, rl.DarkGray) {
		ui.ShowNotesPanel = false
		return
	}

	// draw scroll bar if there are more items than visible
	if len(ui.Notes) > maxVisible {
		scrollBarX := panelX + panelW - 15
		scrollBarY := panelY + 30
		scrollBarH := panelH - 70

		rl.DrawRectangle(scrollBarX, scrollBarY, 10, int32(scrollBarH), rl.Gray)

		// calculate scroll thumb position and size
		if len(ui.Notes) > 0 {
			thumbHeight := int32(float32(scrollBarH) * float32(maxVisible) / float32(len(ui.Notes)))
			if thumbHeight < 10 {
				thumbHeight = 10
			}
			thumbY := scrollBarY + int32(float32(scrollBarH-int32(thumbHeight))*float32(ui.NotesScroll)/float32(len(ui.Notes)-maxVisible))

			rl.DrawRectangle(scrollBarX+1, thumbY, 8, thumbHeight, rl.LightGray)
		}
	}

	// visible list
	startIdx := ui.NotesScroll
	endIdx := startIdx + maxVisible
	if endIdx > len(ui.Notes) {
		endIdx = len(ui.Notes)
	}
	// ui.Notes = listNoteFiles("notes", true)
	fmt.Println("startIDx =", startIdx)
	fmt.Println("endidx =", endIdx)
	fmt.Println("len ui.notes", len(ui.Notes))
	for i := startIdx; i < endIdx; i++ {
		y := panelY + 30 + int32((i-startIdx)*int(entryHeight))
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
		fmt.Println("len ui.notes", len(ui.Notes))

		if DrawButton(entry, panelX+10, y, panelW-30, 20, "white", rl.DarkGreen, rl.Gray, rl.DarkGray) {
			if ui.IsFolderView {
				// Folder clicked
				ui.NotesPath = entry // like "projectA/"
				ui.Notes = listNoteFiles("notes/"+entry, false)
				ui.IsFolderView = false
				ui.NotesScroll = 0
			} else {
				// File clicked
				fullPath := "notes/" + ui.NotesPath + entry
				clearTextGrid()
				loadFileIntoTextGrid(fullPath)
				cursor.reset()
				currentFile = fullPath
				editorStatus = "Loaded Note: " + entry
				ui.ShowNotesPanel = false
			}
		}
	}
}

func DrawDropdown(menu string, x, y int32, ui *UIState) {
	options := []string{"Open...", "Save", "Save As...", "New"}
	for i, opt := range options {
		btnY := y + int32(i*20)
		if DrawButton(opt, x, btnY, 100, 20, "white", rl.DarkGreen, rl.Gray, rl.DarkBlue) {
			switch opt {
			case "Open...":
				ui.ModalOpen = "OpenFile"
				ui.InputBoxes = []*InputBox{
					{
						Rect:     rl.NewRectangle(150, 150, 300, 30),
						Text:     "",
						MaxChars: 64,
					},
				}
			case "Save":
				fmt.Println("Save triggered (implement me!)")
				saveTextGridToFile(currentFile)
				editorStatus = "Saved File!"
			case "Save As...":
				ui.ModalOpen = "SaveAs"
				ui.InputBoxes = []*InputBox{
					{
						Rect:     rl.NewRectangle(150, 150, 300, 30),
						Text:     "",
						MaxChars: 64,
					},
				}
			case "New":
				clearTextGrid()
				cursor.reset()

				printGrid(cursor)
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

		rl.DrawRectangle(modalX, modalY, modalW, modalH, rl.DarkBrown)
		rl.DrawRectangleLines(modalX, modalY, modalW, modalH, rl.Black)

		DrawText("Open File", int(modalX)+10, int(modalY)+10, 10, rl.DrawPixel, "black")

		for _, ib := range ui.InputBoxes {
			ib.Draw()
			ib.HandleInput()
		}

		if DrawButton("Cancel", modalX+modalW-80, modalY+modalH-40, 70, 30, "white", rl.Red, rl.Gray, rl.DarkGray) {
			ui.ModalOpen = ""
		}

		if DrawButton("OK", modalX+modalW-160, modalY+modalH-40, 70, 30, "white", rl.Green, rl.Gray, rl.DarkGray) {
			if len(ui.InputBoxes) > 0 {
				filename := ui.InputBoxes[0].Text
				fmt.Printf("open: %s\n", filename)
				res, err := loadFileIntoTextGrid(filename)
				if err != nil {
					clearTextGrid()
					cursor.reset()
					loadStringIntoTextGrid("File Doesn't Exist!!!")
					// DrawText("File Doesn't Exist.", 0, editorTopPadding, CHAR_IMAGE_WIDTH, rl.DrawPixel, "white")
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

		rl.DrawRectangle(modalX, modalY, modalW, modalH, rl.DarkBrown)
		rl.DrawRectangleLines(modalX, modalY, modalW, modalH, rl.Black)

		DrawText("Save As", int(modalX)+10, int(modalY)+10, 10, rl.DrawPixel, "black")

		for _, ib := range ui.InputBoxes {
			ib.Draw()
			ib.HandleInput()
		}

		if DrawButton("Cancel", modalX+modalW-80, modalY+modalH-40, 70, 30, "white", rl.Red, rl.Gray, rl.DarkGray) {
			ui.ModalOpen = ""
		}

		if DrawButton("OK", modalX+modalW-160, modalY+modalH-40, 70, 30, "white", rl.Green, rl.Gray, rl.DarkGray) {
			if len(ui.InputBoxes) > 0 {
				filename := ui.InputBoxes[0].Text
				err := saveTextGridToFile(filename)
				if err != nil {
					fmt.Println("Error saving file:", err)
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

		rl.DrawRectangle(modalX, modalY, modalW, modalH, rl.DarkBrown)
		rl.DrawRectangleLines(modalX, modalY, modalW, modalH, rl.Black)

		DrawText("Create Note", int(modalX)+10, int(modalY)+10, 10, rl.DrawPixel, "black")

		for _, ib := range ui.InputBoxes {
			ib.Draw()
			ib.HandleInput()
		}

		if DrawButton("Cancel", modalX+modalW-80, modalY+modalH-40, 70, 30, "white", rl.Red, rl.Gray, rl.DarkGray) {
			ui.ModalOpen = ""
		}

		if DrawButton("OK", modalX+modalW-160, modalY+modalH-40, 70, 30, "white", rl.Green, rl.Gray, rl.DarkGray) {
			if len(ui.InputBoxes) > 0 {
				filename := ui.InputBoxes[0].Text

				now := time.Now()
				folderName := now.Format("01-2006")
				baseDir, err := os.Getwd()
				if err != nil {
					fmt.Printf("Error getting current working directory: %v\n", err)
					return
				}

				folderPath := filepath.Join(baseDir, "notes/"+folderName)
				fmt.Printf("Attempting to create folder: %s\n", folderPath)
				_, err = os.Stat(folderPath)
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
					fmt.Println("Error saving file:", err)
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
	rl.DrawRectangle(0, int32(windowHeight)-int32(barHeight), int32(windowWidth), int32(barHeight), rl.DarkGray)
	status := fmt.Sprintf("Ln %d, Col %d Buffer: %s | Status: %s", cursor.y+1, cursor.x+1, currentFile, editorStatus)
	DrawText(status, 5, windowHeight-barHeight+5, CHAR_IMAGE_WIDTH, rl.DrawPixel, "white")
}

func DrawButton(label string, x, y, w, h int32, textColor_ string, pressColor, hoverColor, idleColor rl.Color) bool {
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

	rl.DrawRectangle(x, y, w, h, bgColor)
	rl.DrawRectangleLines(x, y, w, h, rl.Black)

	textSize := rl.MeasureText(label, 12)
	textX := x + (w-int32(textSize))/4
	textY := y + (h-10)/2
	DrawText(label, int(textX), int(textY), CHAR_IMAGE_WIDTH, rl.DrawPixel, textColor_)

	return mouseOver && pressed
}
