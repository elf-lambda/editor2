package main

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func handleEditorInput(cursor *Cursor) {

	if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
		if rl.IsKeyPressed(rl.KeyS) {
			if currentFile != "Untitled" {
				saveTextGridToFile(currentFile)
				fmt.Println("Saved:", currentFile)
				editorStatus = "Saved " + currentFile
			} else {
				ui.ModalOpen = "SaveAs"
				ui.InputBoxes = []*InputBox{
					{
						Rect:     rl.NewRectangle(150, 150, 300, 30),
						Text:     "",
						MaxChars: 64,
					},
				}
			}
		}
	}

	for char := rl.GetCharPressed(); char > 0; char = rl.GetCharPressed() {
		if char >= 32 && char <= 126 {
			cursor.insert(byte(char))
		} else {
			cursor.insert(byte(1))
		}
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		cursor.enter()
	}
	if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
		cursor.backspace()
		// time.Sleep(222 * time.Millisecond)
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		cursor.moveLeft()
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		cursor.moveRight()
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		cursor.moveUp()
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		cursor.moveDown()
	}
	if rl.IsKeyPressed(rl.KeyTab) {
		for i := 0; i < 4; i++ {
			cursor.insert(' ')
		}
	}
}

func main() {
	var file string
	if len(os.Args) > 1 {
		file = os.Args[1]
		loadFileIntoTextGrid(file)
	}

	rl.InitWindow(int32(windowWidth), int32(windowHeight), "Text Editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {

		// --- Input ---
		if ui.ModalOpen == "" {
			handleEditorInput(cursor)
		}

		// --- Rendering ---
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Editor Area
		if ui.ModalOpen == "" {
			for y := 0; y < usedRows; y++ {
				for x := 0; x < editorCols; x++ {
					char := textGrid[y][x]
					if char >= 32 && char <= 126 {
						DrawCharacter(char,
							(x*CHAR_IMAGE_WIDTH)+editorXPadding,
							(y*CHAR_IMAGE_HEIGHT)+editorTopPadding,
							rl.DrawPixel,
							"white")
					} else if char == 0 || char == '\n' {
						continue
					} else {
						DrawCharacter(1,
							(x*CHAR_IMAGE_WIDTH)+editorXPadding,
							(y*CHAR_IMAGE_HEIGHT)+editorTopPadding,
							rl.DrawPixel,
							"white")
						continue
					}

				}
			}

			// Cursor
			DrawCharacter(4,
				(cursor.x*CHAR_IMAGE_WIDTH)+editorXPadding,
				(cursor.y*CHAR_IMAGE_HEIGHT)+editorTopPadding,
				rl.DrawPixel,
				"cyan")
		}

		// UI Layers
		DrawMenuBar(ui)
		DrawStatusBar(*cursor)

		if ui.ModalOpen != "" {
			DrawModal(ui)
		}

		rl.EndDrawing()
	}
}
