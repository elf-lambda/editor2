package main

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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
		if ui.ModalOpen == "" {
			handleEditorInput(cursor)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

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

			// render only visible characters
			for y := startY; y < endY; y++ {
				for x := startX; x < endX; x++ {
					char := textGrid[y][x]
					if char >= 32 && char <= 126 {
						DrawCharacter(char,
							((x-scrollOffsetX)*CHAR_IMAGE_WIDTH)+editorXPadding,
							((y-scrollOffsetY)*CHAR_IMAGE_HEIGHT)+editorTopPadding,
							rl.DrawPixel,
							"white")
					} else if char == 0 || char == '\n' {
						continue
					} else {
						DrawCharacter(1,
							((x-scrollOffsetX)*CHAR_IMAGE_WIDTH)+editorXPadding,
							((y-scrollOffsetY)*CHAR_IMAGE_HEIGHT)+editorTopPadding,
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
					((cursor.y-scrollOffsetY)*CHAR_IMAGE_HEIGHT)+editorTopPadding,
					rl.DrawPixel,
					"cyan")
			}

			// draw scroll indicators
			drawScrollIndicators()
		}

		DrawMenuBar(ui)
		DrawStatusBar(*cursor)

		if ui.ShowNotesPanel {
			DrawNotesPanel(ui)
		}

		if ui.ModalOpen != "" {
			DrawModal(ui)
		}

		rl.EndDrawing()
	}
}
