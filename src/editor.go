package main

import (
	"fmt"
)

const editorPadding int = 15

var windowHeight int = 256
var windowWidth int = 256
var editorCols int = windowWidth / CHAR_IMAGE_WIDTH   // ^ ^ ^
var editorRows int = windowHeight / CHAR_IMAGE_HEIGHT // >
var usedRows int = 1

var textGrid [][]byte = getTextGrid()

type Cursor struct {
	x int // Cols
	y int // Rows
}

func (c *Cursor) enter() {
	if c.y+1 >= editorRows {
		growTextGrid()
	}
	if usedRows >= editorRows {
		growTextGrid()
	}

	// shift lines below down by 1
	for i := usedRows; i > c.y+1; i-- {
		textGrid[i] = textGrid[i-1]
	}

	// insert new empty row after current
	textGrid[c.y+1] = make([]byte, editorCols)

	// move tail of current line to new line
	for i := c.x; i < editorCols; i++ {
		textGrid[c.y+1][i-c.x] = textGrid[c.y][i]
		textGrid[c.y][i] = 0
	}

	textGrid[c.y][c.x] = '*'

	c.y++
	c.x = 0

	usedRows++
}

func (c *Cursor) backspace() {
	// if in line or end
	if c.x > 0 {
		c.x--
		for i := c.x; i < editorCols-1; i++ {
			textGrid[c.y][i] = textGrid[c.y][i+1]
		}
		textGrid[c.y][editorCols-1] = 0
		return
	}

	// if at start of the line
	if c.y > 0 {
		prevY := c.y - 1
		var lastCharX int = -1

		// find the end of the previous line
		for i := editorCols - 1; i >= 0; i-- {
			if textGrid[prevY][i] != 0 {
				lastCharX = i
				break
			}
		}

		// if previous line ends with '*', delete it
		if lastCharX != -1 && textGrid[prevY][lastCharX] == '*' {
			textGrid[prevY][lastCharX] = 0
			lastCharX--
		}

		// compact current line (ignore 0s)
		var compacted []byte
		for i := 0; i < editorCols; i++ {
			if textGrid[c.y][i] != 0 {
				compacted = append(compacted, textGrid[c.y][i])
			}
		}

		// merge into previous line starting after lastCharX
		mergeStart := lastCharX + 1
		if mergeStart+len(compacted) >= editorCols {
			growTextGridCols()
		}
		for i, b := range compacted {
			textGrid[prevY][mergeStart+i] = b
		}

		// shift lines up
		for y := c.y; y < editorRows-1; y++ {
			copy(textGrid[y], textGrid[y+1])
		}

		// clear the last row
		for i := 0; i < editorCols; i++ {
			textGrid[editorRows-1][i] = 0
		}

		c.y--
		c.x = mergeStart + len(compacted)
		if usedRows > 0 {
			usedRows--
		}
	}
}

func (c *Cursor) moveLeft() {
	if c.x > 0 {
		c.x--
	} else if c.y > 0 {
		c.y--
		// go to end of previous line
		for i := editorCols - 1; i >= 0; i-- {
			if textGrid[c.y][i] != 0 {
				c.x = i
				if c.x >= editorCols {
					c.x = editorCols - 1
				}
				return
			}
		}
		c.x = 0
	}
	fmt.Println(c)
}

func (c *Cursor) moveRight() {
	if textGrid[c.y][c.x] == '*' && (textGrid[c.y][c.x+1] == 0 || textGrid[c.y][c.x+1] == '*') {
		return
	}

	if c.x+1 < editorCols && (textGrid[c.y][c.x] != 0) {
		c.x++
		return
	}

	// if at newline marker, move to next line
	if c.x < editorCols && textGrid[c.y][c.x] == '*' && c.y+1 < usedRows {
		c.y++
		c.x = 0
	}
	fmt.Println(c)
}

func (c *Cursor) moveUp() {

	if c.y > 0 {
		c.y--
		c.clampXToLineEnd()
	}
	fmt.Println(c)
}

func (c *Cursor) moveDown() {

	if c.y < usedRows-1 {
		c.y++
		c.clampXToLineEnd()
	}
	fmt.Println(c)
}

func (c *Cursor) clampXToLineEnd() {
	maxX := 0
	for i := 0; i < editorCols; i++ {
		if textGrid[c.y][i] != 0 {
			maxX = i + 1
		}
	}
	if c.x > maxX {
		c.x = maxX
	}
}

func (c *Cursor) insert(char byte) {
	c.checkBounds()

	// if the line is full, push to next line
	// if textGrid[c.y][editorCols-1] != 0 {
	// 	c.enter()
	// }

	// shift characters right from the end to cursor.x
	for i := editorCols - 1; i > c.x; i-- {
		textGrid[c.y][i] = textGrid[c.y][i-1]
	}

	textGrid[c.y][c.x] = char
	c.x++

	// if c.x >= editorCols {
	// 	c.enter()
	// }
}

func (c *Cursor) checkBounds() {
	if c.y >= editorRows {
		growTextGrid()
	}
	if c.x >= editorCols {
		growTextGridCols()
	}
}

func growTextGrid() {
	newRows := editorRows * 2
	newGrid := make([][]byte, newRows)

	// copy old rows
	for i := 0; i < editorRows; i++ {
		newGrid[i] = textGrid[i]
	}

	// initialize new empty rows
	for i := editorRows; i < newRows; i++ {
		newGrid[i] = make([]byte, editorCols)
	}

	textGrid = newGrid
	editorRows = newRows
	fmt.Printf("Text grid expanded to %d rows\n", editorRows)
}

func growTextGridCols() {
	newCols := editorCols * 2
	for i := 0; i < editorRows; i++ {
		oldRow := textGrid[i]
		newRow := make([]byte, newCols)
		copy(newRow, oldRow)
		textGrid[i] = newRow
	}
	editorCols = newCols
	fmt.Printf("Text grid expanded to %d cols\n", editorCols)
}

func getTextGrid() [][]byte {
	fmt.Println(editorRows, editorCols)
	var textgrid = make([][]byte, editorRows)
	for i := range textgrid {
		textgrid[i] = make([]byte, editorCols)
	}
	return textgrid
}

func printGrid(cursor *Cursor) {
	fmt.Println("------------")
	for r, _ := range textGrid {
		for c, colValue := range textGrid[r] {
			if r == cursor.y && c == cursor.x {
				fmt.Printf("[@]")
				continue
			}
			if colValue == 0 {
				fmt.Printf("[ ]")
				continue
			}
			fmt.Printf("[%c]", colValue)
		}
		fmt.Println()
	}
	fmt.Println("------------")
}

func testCursor() {
	cursor := &Cursor{}
	cursor.insert('H')
	cursor.insert('i')
	cursor.insert('1')

	printGrid(cursor)
	cursor.enter()
	printGrid(cursor)
	cursor.moveUp()
	printGrid(cursor)
	cursor.moveDown()
	cursor.moveDown()
	cursor.moveRight()
	printGrid(cursor)
	cursor.moveUp()
	printGrid(cursor)
	cursor.enter()
	cursor.moveRight()
	printGrid(cursor)
	cursor.enter()
	printGrid(cursor)
	cursor.moveRight()
	cursor.moveRight()

	cursor.moveRight()

	cursor.moveRight()
	cursor.moveRight()
	printGrid(cursor)
	cursor.enter()
	printGrid(cursor)
}
