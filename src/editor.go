package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const editorXPadding int = 5
const editorYPadding int = 5
const editorTopPadding int = 30
const editorBottomPadding int = 20

const windowHeight int = 460
const windowWidth int = 640

var editorCols int = windowWidth / CHAR_IMAGE_WIDTH   // ^ ^ ^
var editorRows int = windowHeight / CHAR_IMAGE_HEIGHT // >
var usedRows int = 1

var currentFile string = "Untitled"
var textGrid [][]byte = getTextGrid(editorRows, editorCols)

var cursor = &Cursor{}
var ui = &UIState{
	CurrentView: "editor",
}

var editorStatus string = ""
var editorClipboard string

// TODO: Refactor Notes to use the new FileEntry structs and functions
type FileEntry struct {
	Name     string
	IsFolder bool
}

func listAllFiles(dir string) []FileEntry {
	if dir == "" {
		dir = "."
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var entries []FileEntry

	// add folders first
	for _, f := range files {
		if f.IsDir() {
			entries = append(entries, FileEntry{
				Name:     f.Name(),
				IsFolder: true,
			})
		}
	}

	// files last
	for _, f := range files {
		if !f.IsDir() {
			entries = append(entries, FileEntry{
				Name:     f.Name(),
				IsFolder: false,
			})
		}
	}

	return entries
}

func ensureGridCapacityForPaste(startX, startY int, content string) {
	lines := strings.Split(content, "\n")
	requiredRows := startY + len(lines)
	maxLineLen := 0
	for _, line := range lines {
		if len(line) > maxLineLen {
			maxLineLen = len(line)
		}
	}
	requiredCols := startX + maxLineLen

	// grow rows if needed
	for len(textGrid) < requiredRows {
		textGrid = append(textGrid, make([]byte, editorCols))
		editorRows++
	}

	// grow columns if needed
	if requiredCols > editorCols {
		for i := range textGrid {
			newRow := make([]byte, requiredCols)
			copy(newRow, textGrid[i])
			textGrid[i] = newRow
		}
		editorCols = requiredCols
	}
}

func insertStringAtCursor(s string) {
	ensureCursorVisible(cursor)
	for i := 0; i < len(s); i++ {
		cursor.checkBounds()
		ch := s[i]
		if ch == '\n' {
			cursor.enter()
		} else if ch == '\t' {
			cursor.insert(' ')
			cursor.insert(' ')
			cursor.insert(' ')
			cursor.insert(' ')
		} else {
			cursor.insert(ch)
		}
	}
	// cursor.enter()
}
func insertStringAt(x, y int, s string) {
	cursor.x = x
	cursor.y = y

	insertStringAtCursor(s)

	usedRows = max(usedRows, cursor.y+1)
}

func listNoteFiles(dir string, foldersOnly bool) []string {
	// Old --- to refactor
	os.MkdirAll("/home/void/notes", os.ModePerm)

	files, err := os.ReadDir("/home/void/" + dir)
	if err != nil {
		fmt.Println("error reading dir listnotefiles")
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

	slices.Reverse(entries)
	fmt.Println(entries)
	return entries
}

func clampName(in string, width int32, charWidth int) string {
	maxChars := int(width) / charWidth

	if len(in)*charWidth > int(width) || len(in) > maxChars {
		if maxChars < 3 {
			return "..."
		}

		charsToTake := maxChars

		if len(in) <= charsToTake {
			return in
		}
		tmp := ""
		for i := 0; i < charsToTake && i < len(in); i++ {
			tmp += string(in[i])
		}

		return tmp + "..."
	}

	return in
}
func deleteFile(path string) {
	fileName := path
	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("Attempting to delete " + fileName)
		copyFile(fileName, fileName+"_backup")
		err = os.Remove(fileName)
		if err != nil {
			fmt.Println("Failed to delete file!")
			return
		}
		fmt.Println("Deleted file: " + fileName)
	}
}

func copyFile(src string, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTextGrid() {
	var editorCols_ int = windowWidth / CHAR_IMAGE_WIDTH   // ^ ^ ^
	var editorRows_ int = windowHeight / CHAR_IMAGE_HEIGHT // >
	fmt.Println("Resizing textGrid to Cols:", editorCols_, "Rows:", editorRows_)
	textGrid = getTextGrid(editorRows_, editorCols_)
	editorRows = editorRows_
	editorCols = editorCols_
	usedRows = 1
	cursor.reset()
	editorStatus = "New Buffer Created"
	currentFile = "Untitled"
}

func loadFileIntoTextGrid(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	clearTextGrid()
	cursor.reset()
	count := 0

	reader := bufio.NewReader(file)
	for {
		char, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return count, err
		}

		cursor.checkBounds()
		if char == '\n' {
			cursor.enter()
			continue
		}
		if char == '\t' {
			cursor.insert(' ')
			cursor.insert(' ')
			cursor.insert(' ')
			cursor.insert(' ')
			continue
		}

		cursor.insert(char)
	}
	currentFile = path
	fmt.Println("Loaded file: ", path)
	// printGrid(cursor)
	return count, nil
}

func loadStringIntoTextGrid(content string) {
	clearTextGrid()
	cursor.reset()

	for i := 0; i < len(content); i++ {
		char := content[i]

		if char == '\n' {
			cursor.enter()
			continue
		}

		cursor.insert(char)
	}

	usedRows = cursor.y + 1
}

func saveTextGridToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for y := 0; y < usedRows; y++ {

		// write chars in the row
		for x := 0; x < editorCols; x++ {
			char := textGrid[y][x]
			if char == 0 {
				continue
			}
			writer.WriteByte(char)
		}

	}

	return writer.Flush()
}

// ------------------------------------------------------------------------------------

type Selection struct {
	Active bool
	StartX int
	StartY int
	EndX   int
	EndY   int
}

var selection Selection

type Cursor struct {
	x int // Cols
	y int // Rows
}

func (c *Cursor) reset() {
	c.x = 0
	c.y = 0
}

func (c *Cursor) String() string {
	return fmt.Sprintf("Cursor[%d, %d]", c.x, c.y)
}

func (c *Cursor) MoveToClick(x, y int) {
	// if cell has character, just move there
	if textGrid[y][x] != 0 {
		c.x = x
		c.y = y
		return
	}

	// look left in the same line
	for i := x - 1; i >= 0; i-- {
		if textGrid[y][i] != 0 {
			c.x = i
			c.y = y
			return
		}
	}

	// look upward
	for j := y - 1; j >= 0; j-- {
		for i := editorCols - 1; i >= 0; i-- {
			if textGrid[j][i] != 0 {
				c.x = i
				c.y = j
				return
			}
		}
	}

	// fallback, no idea if this can be reached :)
	c.reset()
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

	textGrid[c.y][c.x] = '\n'

	c.y++
	c.x = 0

	usedRows++
}

func (c *Cursor) backspace() {
	if selection.Active {
		// if we have a selection

		startX, startY := selection.StartX, selection.StartY
		endX, endY := selection.EndX, selection.EndY

		// normalize selection
		if startY > endY || (startY == endY && startX > endX) {
			startX, endX = endX, startX
			startY, endY = endY, startY
		}

		// move cursor to end of selection
		c.x = endX
		c.y = endY

		// repeatedly call single-character backspace until we reach selection start
		for c.y > startY || (c.y == startY && c.x > startX) {
			c.backspaceSingle()
		}
		// recalculate usedRows after deletion
		usedRows = 0
		for y := editorRows - 1; y >= 0; y-- {
			nonEmpty := false
			for x := 0; x < editorCols; x++ {
				if textGrid[y][x] != 0 {
					nonEmpty = true
					break
				}
			}
			if nonEmpty {
				usedRows = y + 1
				break
			}
		}
		if usedRows == 0 {
			usedRows = 1
		}
		// TODO: Maybe recalculate columns aswell?

		selection.Active = false
		ensureCursorVisible(c)
		return
	}

	// regular backspace if no selection
	c.backspaceSingle()
	ensureCursorVisible(c)
}

func (c *Cursor) backspaceSingle() {
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

		// if previous line ends with '\n', delete it
		if lastCharX != -1 && textGrid[prevY][lastCharX] == '\n' {
			textGrid[prevY][lastCharX] = 0
			lastCharX--
		}

		// compact current line
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
	if textGrid[c.y][c.x] == '\n' && (textGrid[c.y][c.x+1] == 0 || textGrid[c.y][c.x+1] == '\n') {
		// c.x++
		c.y++
		c.x = 0
		return
	}

	if c.x+1 < editorCols && (textGrid[c.y][c.x] != 0) {
		c.x++
		return
	}

	// if at newline marker, move to next line
	if c.x < editorCols && textGrid[c.y][c.x] == '\n' && c.y+1 < usedRows {
		c.y++
		c.x = 0
	}
	fmt.Println(c)
}

func (c *Cursor) moveUp() {

	if c.y > 0 {
		c.y--
		if c.x > editorCols {
			c.clampXToLineEnd()
		}
		if textGrid[c.y][c.x] == 0 {
			c.clampXToLineEnd()
		}
	}
	fmt.Println(c)
}

func (c *Cursor) moveDown() {

	if c.y < usedRows-1 {
		c.y++
		if c.x > editorCols {
			c.clampXToLineEnd()
		}
		if textGrid[c.y][c.x] == 0 {
			c.clampXToLineEnd()
		}
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
		c.x = maxX - 1
	}
	c.x = maxX - 1
}

func (c *Cursor) insert(char byte) {
	c.checkBounds()

	// shift characters right from the end to cursor.x
	for i := editorCols - 1; i > c.x; i-- {
		textGrid[c.y][i] = textGrid[c.y][i-1]
	}

	textGrid[c.y][c.x] = char
	c.x++

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

func getTextGrid(rows, cols int) [][]byte {
	fmt.Println(rows, cols)
	var textgrid = make([][]byte, rows)
	for i := range textgrid {
		textgrid[i] = make([]byte, cols)
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
