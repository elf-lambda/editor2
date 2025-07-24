package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type UIState struct {
	ActiveMenu string
	ModalOpen  string
	InputBoxes []*InputBox
}

type InputBox struct {
	Rect     rl.Rectangle
	Text     string
	Focused  bool
	MaxChars int
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

	if DrawButton("File", 0, 0, 60, int32(menuHeight), "white", rl.DarkGreen, rl.Gray, rl.DarkBlue) {
		if ui.ActiveMenu == "File" {
			ui.ActiveMenu = ""
		} else {
			ui.ActiveMenu = "File"
		}
	}

	if ui.ActiveMenu == "File" {
		DrawDropdown("File", 0, int32(menuHeight), ui)
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
				editorStatus = "New Buffer Created"
				currentFile = "Untitled"
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

	textSize := rl.MeasureText(label, 14)
	textX := x + (w-int32(textSize))/2
	textY := y + (h-10)/2
	DrawText(label, int(textX)-10, int(textY), CHAR_IMAGE_WIDTH, rl.DrawPixel, textColor_)

	return mouseOver && pressed
}
