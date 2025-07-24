package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var scrollOffsetX int = 0
var scrollOffsetY int = 0

func getVisibleRows() int {
	return (windowHeight - editorTopPadding - editorBottomPadding) / CHAR_IMAGE_HEIGHT
}

func getVisibleCols() int {
	return (windowWidth - editorXPadding*2) / CHAR_IMAGE_WIDTH
}

func getRowWidth(row int) int {
	if row >= usedRows {
		return 0
	}
	maxX := 0
	for i := 0; i < editorCols; i++ {
		if textGrid[row][i] != 0 {
			maxX = i + 1
		}
	}
	return maxX
}

func getMaxContentWidth() int {
	maxWidth := 0
	for y := 0; y < usedRows; y++ {
		width := getRowWidth(y)
		if width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth == 0 {
		maxWidth = 1
	}
	return maxWidth
}

func ensureCursorVisible(cursor *Cursor) {
	visibleRows := getVisibleRows()
	visibleCols := getVisibleCols()

	// vertical scrolling
	if cursor.y < scrollOffsetY {
		scrollOffsetY = cursor.y
	} else if cursor.y >= scrollOffsetY+visibleRows {
		scrollOffsetY = cursor.y - visibleRows + 1
	}

	// horizontal scrolling
	if cursor.x < scrollOffsetX {
		scrollOffsetX = cursor.x
	} else if cursor.x >= scrollOffsetX+visibleCols {
		scrollOffsetX = cursor.x - visibleCols + 1
	}

	// check to not scroll beyond content
	maxScrollY := usedRows - visibleRows
	if maxScrollY < 0 {
		maxScrollY = 0
	}
	if scrollOffsetY > maxScrollY {
		scrollOffsetY = maxScrollY
	}
	if scrollOffsetY < 0 {
		scrollOffsetY = 0
	}

	maxScrollX := getMaxContentWidth() - visibleCols
	if maxScrollX < 0 {
		maxScrollX = 0
	}
	if scrollOffsetX > maxScrollX {
		scrollOffsetX = maxScrollX
	}
	if scrollOffsetX < 0 {
		scrollOffsetX = 0
	}
}

func handleEditorInput(cursor *Cursor) {
	mouseWheel := rl.GetMouseWheelMove()
	if mouseWheel != 0 {
		if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
			// handle horizontal scrolling with Shift+Scroll
			scrollOffsetX -= int(mouseWheel * 5) // Scroll 5 chars at a time
			maxScrollX := getMaxContentWidth() - getVisibleCols()
			if maxScrollX < 0 {
				maxScrollX = 0
			}
			if scrollOffsetX < 0 {
				scrollOffsetX = 0
			}
			if scrollOffsetX > maxScrollX {
				scrollOffsetX = maxScrollX
			}
		} else {
			// vertical scrolling
			scrollOffsetY -= int(mouseWheel * 3) // 3 lines at a time
			maxScrollY := usedRows - getVisibleRows()
			if maxScrollY < 0 {
				maxScrollY = 0
			}
			if scrollOffsetY < 0 {
				scrollOffsetY = 0
			}
			if scrollOffsetY > maxScrollY {
				scrollOffsetY = maxScrollY
			}
		}
	}

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
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		cursor.enter()
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
		cursor.backspace()
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressedRepeat(rl.KeyLeft) {
		cursor.moveLeft()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressedRepeat(rl.KeyRight) {
		cursor.moveRight()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressedRepeat(rl.KeyUp) {
		cursor.moveUp()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressedRepeat(rl.KeyDown) {
		cursor.moveDown()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyTab) {
		for i := 0; i < 4; i++ {
			cursor.insert(' ')
		}
		ensureCursorVisible(cursor)
	}

	// ----- gen`1`

	// Page Up/Down for faster scrolling
	if rl.IsKeyPressed(rl.KeyPageUp) {
		visibleRows := getVisibleRows()
		cursor.y -= visibleRows
		if cursor.y < 0 {
			cursor.y = 0
		}
		cursor.clampXToLineEnd()
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyPageDown) {
		visibleRows := getVisibleRows()
		cursor.y += visibleRows
		if cursor.y >= usedRows {
			cursor.y = usedRows - 1
		}
		cursor.clampXToLineEnd()
		ensureCursorVisible(cursor)
	}

	// Home/End keys
	if rl.IsKeyPressed(rl.KeyHome) {
		if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
			// Ctrl+Home: Go to beginning of document
			cursor.x = 0
			cursor.y = 0
		} else {
			// Home: Go to beginning of line
			cursor.x = 0
		}
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyEnd) {
		if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
			// Ctrl+End: Go to end of document
			cursor.y = usedRows - 1
			cursor.clampXToLineEnd()
		} else {
			// End: Go to end of line
			cursor.clampXToLineEnd()
		}
		ensureCursorVisible(cursor)
	}

	// ----- gen`1`
}

func drawScrollIndicators() {
	visibleRows := getVisibleRows()
	visibleCols := getVisibleCols()
	maxContentWidth := getMaxContentWidth()

	// vertical scrollbar
	if usedRows > visibleRows {
		scrollBarX := int32(windowWidth - 10)
		scrollBarY := int32(editorTopPadding)
		scrollBarH := int32(windowHeight - editorTopPadding - editorBottomPadding)

		rl.DrawRectangle(scrollBarX, scrollBarY, 8, scrollBarH, rl.DarkGray)

		thumbHeight := int32(float32(scrollBarH) * float32(visibleRows) / float32(usedRows))
		if thumbHeight < 10 {
			thumbHeight = 10
		}

		maxScrollY := usedRows - visibleRows
		if maxScrollY > 0 {
			thumbY := scrollBarY + int32(float32(scrollBarH-thumbHeight)*float32(scrollOffsetY)/float32(maxScrollY))

			rl.DrawRectangle(scrollBarX+1, thumbY, 6, thumbHeight, rl.Gray)
		}
	}

	// horizontal scroll bar
	if maxContentWidth > visibleCols {
		scrollBarX := int32(editorXPadding)
		scrollBarY := int32(windowHeight - (editorBottomPadding * 2) + 10)
		scrollBarW := int32(windowWidth - editorXPadding*2 - 15)

		rl.DrawRectangle(scrollBarX, scrollBarY, scrollBarW, 6, rl.DarkGray)

		thumbWidth := int32(float32(scrollBarW) * float32(visibleCols) / float32(maxContentWidth))
		if thumbWidth < 10 {
			thumbWidth = 10
		}

		maxScrollX := maxContentWidth - visibleCols
		if maxScrollX > 0 {
			thumbX := scrollBarX + int32(float32(scrollBarW-thumbWidth)*float32(scrollOffsetX)/float32(maxScrollX))
			rl.DrawRectangle(thumbX, scrollBarY+1, thumbWidth, 4, rl.Gray)
		}
	}
}
