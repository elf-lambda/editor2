package main

import (
	"image/color"
)

// Draws a character at the specified coordinates
func DrawCharacter(c byte, startX, startY int, fn func(xIn, yIn int32, col color.RGBA)) {
	fontChar := fontCharacters[c]

	// Iterate over the character pixel data
	for y := 0; y < fontChar.height; y++ {
		for x := 0; x < fontChar.width; x++ {
			// Calculate the image coordinates for the pixel
			imgX := startX + x
			imgY := startY + y

			// Get the index of the pixel data for this character
			charIndex := y*fontChar.width + x

			// Get the pixel color data for this index
			col := fontChar.data[charIndex]
			if col == 0x000000 {
				continue
			}
			r := byte((col >> 16) & 0xFF) // Extract the red component
			g := byte((col >> 8) & 0xFF)  // Extract the green component
			b := byte(col & 0xFF)         // Extract the blue component

			// Set the color and draw the pixel at the correct position
			fn(int32(imgX), int32(imgY), color.RGBA{r, g, b, 255})
		}
	}
}

// Draw text at specified coordinates
func DrawText(input string, posX int, posY int, charWidth int, fn func(xIn, yIn int32, col color.RGBA)) {
	for i, v := range input {
		DrawCharacter(byte(v), posX+(charWidth*i), posY, fn)
	}
}
