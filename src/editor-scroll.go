package main

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
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
		if textGrid[row][i] == '\n' {
			return i + 1
		}
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
			// ensureCursorVisible(cursor)
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
			// ensureCursorVisible(cursor)
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

	if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
		// Exit
		if rl.IsKeyPressed(rl.KeyQ) {

			os.Exit(1)
		}

		// Copy
		if rl.IsKeyPressed(rl.KeyC) && selection.Active {
			editorClipboard = getSelectedText()
			// rl.SetClipboardText(editorClipboard)
			clipboard.Write(clipboard.FmtText, []byte(editorClipboard))
			selection.reset()
		}

		// Paste
		if rl.IsKeyPressed(rl.KeyV) && editorClipboard != "" {
			ensureGridCapacityForPaste(cursor.x, cursor.y, editorClipboard)
			undoStack = append(undoStack, takeSnapshot())
			redoStack = nil

			insertStringAtCursor(editorClipboard)
			selection.Active = false
			ensureCursorVisible(cursor)
		}

		if (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) && rl.IsKeyPressed(rl.KeyA) {
			selection.Active = true
			selection.StartX = 0
			selection.StartY = 0
			selection.EndY = editorRows - 1
			selection.EndX = editorCols - 1
		}

	}
	// selection arrows
	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressedRepeat(rl.KeyLeft) {
			selection.Active = true
			if !selection.ArrowSelect {
				selection.StartX = cursor.x
				selection.StartY = cursor.y
				selection.ArrowSelect = true
			}
			selection.EndX = cursor.x - 1
			selection.EndY = cursor.y
			cursor.moveLeft()
			return

		}
		if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressedRepeat(rl.KeyRight) {
			selection.Active = true
			if !selection.ArrowSelect {
				selection.StartX = cursor.x
				selection.StartY = cursor.y
				selection.ArrowSelect = true
			}
			selection.EndX = cursor.x + 1
			selection.EndY = cursor.y
			cursor.moveRight()
			return
		}
		if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressedRepeat(rl.KeyUp) {
			selection.Active = true
			if !selection.ArrowSelect {
				selection.StartX = cursor.x
				selection.StartY = cursor.y
				selection.ArrowSelect = true
			}
			cursor.moveUp()
			selection.EndX = cursor.x
			selection.EndY = cursor.y
			return
		}
		if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressedRepeat(rl.KeyDown) {
			selection.Active = true
			if !selection.ArrowSelect {
				selection.StartX = cursor.x
				selection.StartY = cursor.y
				selection.ArrowSelect = true
			}
			cursor.moveDown()
			selection.EndX = cursor.x
			selection.EndY = cursor.y
			return
		}

	}

	for char := rl.GetCharPressed(); char > 0; char = rl.GetCharPressed() {
		if char >= 32 && char <= 126 {
			undoStack = append(undoStack, takeSnapshot())
			redoStack = nil
			cursor.insert(byte(char))
		} else {
			undoStack = append(undoStack, takeSnapshot())
			redoStack = nil
			cursor.insert(byte(1))
		}
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		cursor.enter()
		selection.reset()
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
		cursor.backspace()
		selection.reset()
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressedRepeat(rl.KeyLeft) {
		cursor.moveLeft()
		selection.reset()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressedRepeat(rl.KeyRight) {
		cursor.moveRight()
		selection.reset()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressedRepeat(rl.KeyUp) {
		cursor.moveUp()
		selection.reset()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressedRepeat(rl.KeyDown) {
		cursor.moveDown()
		selection.reset()
		ensureCursorVisible(cursor)
		// time.Sleep(33 * time.Millisecond)
	}

	if rl.IsKeyPressed(rl.KeyTab) {
		for i := 0; i < 4; i++ {
			cursor.insert(' ')
		}
		ensureCursorVisible(cursor)
	}

	if rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyPressed(rl.KeyZ) {
		undo()
	}
	if rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyDown(rl.KeyLeftShift) && rl.IsKeyPressed(rl.KeyZ) {
		redo()
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
			cursor.reset()
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

func getSelectedText() string {
	if !selection.Active {
		return ""
	}

	startX, startY := selection.StartX, selection.StartY
	endX, endY := selection.EndX, selection.EndY

	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	result := ""

	for y := startY; y <= endY; y++ {
		lineStart := 0
		lineEnd := editorCols - 1

		if y == startY {
			lineStart = startX
		}
		if y == endY {
			lineEnd = endX
		}

		for x := lineStart; x <= lineEnd; x++ {
			ch := textGrid[y][x]
			if ch == 0 {
				break
			}
			result += string(ch)
		}

	}
	fmt.Println("getSelectedText: ", result)
	return result
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
