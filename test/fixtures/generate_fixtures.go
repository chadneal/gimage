// Package main generates test fixture images
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func main() {
	// Create a simple 800x600 test image with colored sections
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill with different colors in quadrants
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			var c color.RGBA
			if x < 400 && y < 300 {
				c = color.RGBA{255, 0, 0, 255} // Red top-left
			} else if x >= 400 && y < 300 {
				c = color.RGBA{0, 255, 0, 255} // Green top-right
			} else if x < 400 && y >= 300 {
				c = color.RGBA{0, 0, 255, 255} // Blue bottom-left
			} else {
				c = color.RGBA{255, 255, 0, 255} // Yellow bottom-right
			}
			img.Set(x, y, c)
		}
	}

	// Save test image
	f, err := os.Create("test_image.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}

	log.Println("Created test_image.png (800x600)")

	// Create a small test image
	smallImg := image.NewRGBA(image.Rect(0, 0, 200, 150))
	for y := 0; y < 150; y++ {
		for x := 0; x < 200; x++ {
			c := color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255}
			smallImg.Set(x, y, c)
		}
	}

	f2, err := os.Create("small_test.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()

	if err := png.Encode(f2, smallImg); err != nil {
		log.Fatal(err)
	}

	log.Println("Created small_test.png (200x150)")
}
