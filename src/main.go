package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	cursor := &Cursor{}
	rl.InitWindow(int32(windowWidth), int32(windowHeight), "Text Editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		printGrid(cursor)

		for char := rl.GetCharPressed(); char > 0; char = rl.GetCharPressed() {
			if char >= 32 && char <= 126 { // printable ASCII
				cursor.insert(byte(char))
			}
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			cursor.enter()
		}
		if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
			cursor.backspace()
			time.Sleep(33)
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
			cursor.insert(' ')
			cursor.insert(' ')
			cursor.insert(' ')
			cursor.insert(' ')
		}

		// draw everything
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		for y := 0; y < usedRows; y++ {
			for x := 0; x < editorCols; x++ {
				char := textGrid[y][x]
				if char != 0 {
					DrawCharacter(char, x*CHAR_IMAGE_WIDTH, y*CHAR_IMAGE_HEIGHT, rl.DrawPixel)
				}
			}
		}

		// draw cursor
		DrawCharacter(4, cursor.x*CHAR_IMAGE_WIDTH, cursor.y*CHAR_IMAGE_HEIGHT, rl.DrawPixel)

		rl.EndDrawing()
	}
}
