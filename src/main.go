package main

import (
	"fmt"
	"os"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	var file string
	if len(os.Args) > 1 {
		file = os.Args[1]
		code, err := loadFileIntoTextGrid(file)
		if code == -1 {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	rl.InitWindow(int32(windowWidth), int32(windowHeight), "Text Editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	editorClipboard = rl.GetClipboardText()
	var clipboardMutex sync.Mutex

	for !rl.WindowShouldClose() {
		if ui.ModalOpen == "" {
			handleEditorInput(cursor)

			// handle mouse click to reposition cursor
			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				// start selection
				mouseX := rl.GetMouseX()
				mouseY := rl.GetMouseY()

				gridX := (int(mouseX) - editorXPadding) / CHAR_IMAGE_WIDTH
				gridY := (int(mouseY) - editorTopPadding) / CHAR_IMAGE_HEIGHT

				gridX += scrollOffsetX
				gridY += scrollOffsetY

				if gridX >= 0 && gridX < editorCols && gridY >= 0 && gridY < editorRows {
					cursor.MoveToClick(gridX, gridY)

					selection.Active = true
					selection.StartX = gridX
					selection.StartY = gridY
					selection.EndX = gridX
					selection.EndY = gridY
				}
			}

			if rl.IsMouseButtonDown(rl.MouseLeftButton) && selection.Active {
				mouseX := rl.GetMouseX()
				mouseY := rl.GetMouseY()

				gridX := (int(mouseX) - editorXPadding) / CHAR_IMAGE_WIDTH
				gridY := (int(mouseY) - editorTopPadding) / CHAR_IMAGE_HEIGHT

				gridX += scrollOffsetX
				gridY += scrollOffsetY

				if gridX >= 0 && gridX < editorCols && gridY >= 0 && gridY < editorRows {
					selection.EndX = gridX
					selection.EndY = gridY
				}
			}

			if rl.IsMouseButtonReleased(rl.MouseLeftButton) && selection.Active {
				if selection.StartX == selection.EndX && selection.StartY == selection.EndY {
					selection.Active = false // No drag = no selection
				}
			}

		}

		rl.BeginDrawing()
		rl.ClearBackground(ModernDarkBg)

		if ui.ModalOpen == "" {
			visibleRows := getVisibleRows()
			visibleCols := getVisibleCols()

			startY := scrollOffsetY
			endY := scrollOffsetY + visibleRows
			if endY > usedRows {
				endY = usedRows
			}

			startX := scrollOffsetX
			endX := scrollOffsetX + visibleCols
			maxContentWidth := getMaxContentWidth()
			if endX > maxContentWidth {
				endX = maxContentWidth
			}

			// line highlight
			rl.DrawRectangle(
				int32(editorXPadding),
				int32((cursor.y-scrollOffsetY)*CHAR_IMAGE_HEIGHT+editorTopPadding+editorYPadding),
				int32(windowWidth-20), CHAR_IMAGE_HEIGHT, rl.NewColor(80, 82, 122, 100))

			// render only visible characters
			for y := startY; y < endY; y++ {
				for x := startX; x < endX; x++ {
					screenX := ((x - scrollOffsetX) * CHAR_IMAGE_WIDTH) + editorXPadding
					screenY := ((y - scrollOffsetY) * CHAR_IMAGE_HEIGHT) + editorTopPadding

					// dodge 0 and \n for selection aswell
					if textGrid[y][x] == 0 || textGrid[y][x] == '\n' {
						continue
					}
					// draw selection
					if isCellSelected(x, y) {
						rl.DrawRectangle(int32(screenX), int32(screenY)+int32(editorYPadding), CHAR_IMAGE_WIDTH, CHAR_IMAGE_HEIGHT, ModernLight)
					}

					char := textGrid[y][x]
					if char >= 32 && char <= 126 {
						DrawCharacter(char,
							((x-scrollOffsetX)*CHAR_IMAGE_WIDTH)+editorXPadding,
							((y-scrollOffsetY)*CHAR_IMAGE_HEIGHT)+editorTopPadding+editorYPadding,
							rl.DrawPixel,
							"white")
					} else if char == 0 || char == '\n' {
						continue
					} else {
						// ' ' space char
						DrawCharacter(1,
							((x-scrollOffsetX)*CHAR_IMAGE_WIDTH)+editorXPadding,
							((y-scrollOffsetY)*CHAR_IMAGE_HEIGHT)+editorTopPadding+editorYPadding,
							rl.DrawPixel,
							"white")
						continue
					}
				}
			}

			// render cursor only if it's visible
			if cursor.y >= scrollOffsetY && cursor.y < scrollOffsetY+visibleRows &&
				cursor.x >= scrollOffsetX && cursor.x < scrollOffsetX+visibleCols {
				DrawCharacter(4,
					((cursor.x-scrollOffsetX)*CHAR_IMAGE_WIDTH)+editorXPadding,
					((cursor.y-scrollOffsetY)*CHAR_IMAGE_HEIGHT)+editorTopPadding+editorYPadding,
					rl.DrawPixel,
					"red")
			}

			// draw scroll indicators
			drawScrollIndicators()
		}

		DrawMenuBar(ui)
		DrawStatusBar(*cursor)
		// fmt.Println("clipboard:", editorClipboard)
		clipboardMutex.Lock()
		tmp := rl.GetClipboardText()
		// fmt.Println("tmp clipboard: ", tmp)
		if tmp != editorClipboard && tmp != "" {
			editorClipboard = tmp
			fmt.Println("Clipboard updated from system (window focused):", editorClipboard)
		}
		clipboardMutex.Unlock()

		if ui.ShowNotesPanel {
			DrawNotesPanel(ui)
		}
		if ui.ShowFilePicker {
			DrawFilePickerPanel(ui)
		}
		if ui.ModalOpen != "" {
			DrawModal(ui)
		}

		rl.EndDrawing()
	}
}
