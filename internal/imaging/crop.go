// Package imaging provides image processing operations using pure Go.
package imaging

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

// CropImage crops a rectangular region from an image.
//
// Parameters:
//   - inputPath: path to input image
//   - outputPath: path to save cropped image
//   - x, y: top-left corner coordinates of the crop region
//   - width, height: dimensions of crop region
//
// Returns error if:
//   - input file does not exist or cannot be read
//   - crop region is outside image bounds
//   - dimensions are not positive
//   - output cannot be written
func CropImage(inputPath, outputPath string, x, y, width, height int) error {
	// Validate dimensions
	if width <= 0 {
		return fmt.Errorf("width must be positive, got %d", width)
	}
	if height <= 0 {
		return fmt.Errorf("height must be positive, got %d", height)
	}

	// Validate coordinates
	if x < 0 {
		return fmt.Errorf("x coordinate must be non-negative, got %d", x)
	}
	if y < 0 {
		return fmt.Errorf("y coordinate must be non-negative, got %d", y)
	}

	// Load the input image
	img, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", inputPath, err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop region is within image bounds
	if x >= imgWidth {
		return fmt.Errorf("x coordinate %d is outside image width %d", x, imgWidth)
	}
	if y >= imgHeight {
		return fmt.Errorf("y coordinate %d is outside image height %d", y, imgHeight)
	}
	if x+width > imgWidth {
		return fmt.Errorf("crop region (x=%d + width=%d = %d) exceeds image width %d", x, width, x+width, imgWidth)
	}
	if y+height > imgHeight {
		return fmt.Errorf("crop region (y=%d + height=%d = %d) exceeds image height %d", y, height, y+height, imgHeight)
	}

	// Create crop rectangle
	cropRect := image.Rect(x, y, x+width, y+height)

	// Perform the crop
	cropped := imaging.Crop(img, cropRect)

	// Save the cropped image
	if err := imaging.Save(cropped, outputPath); err != nil {
		return fmt.Errorf("failed to save cropped image to %s: %w", outputPath, err)
	}

	return nil
}

// CropCenter crops a region from the center of the image.
//
// Parameters:
//   - inputPath: path to input image
//   - outputPath: path to save cropped image
//   - width, height: dimensions of crop region
//
// The crop region will be centered on the image. If the requested dimensions
// are larger than the image, an error is returned.
func CropCenter(inputPath, outputPath string, width, height int) error {
	// Validate dimensions
	if width <= 0 {
		return fmt.Errorf("width must be positive, got %d", width)
	}
	if height <= 0 {
		return fmt.Errorf("height must be positive, got %d", height)
	}

	// Load the input image
	img, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", inputPath, err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop region fits within image
	if width > imgWidth {
		return fmt.Errorf("crop width %d exceeds image width %d", width, imgWidth)
	}
	if height > imgHeight {
		return fmt.Errorf("crop height %d exceeds image height %d", height, imgHeight)
	}

	// Use imaging.CropCenter which handles center calculation
	cropped := imaging.CropCenter(img, width, height)

	// Save the cropped image
	if err := imaging.Save(cropped, outputPath); err != nil {
		return fmt.Errorf("failed to save cropped image to %s: %w", outputPath, err)
	}

	return nil
}

// CropAnchor crops a region with a specific anchor point.
//
// Parameters:
//   - inputPath: path to input image
//   - outputPath: path to save cropped image
//   - width, height: dimensions of crop region
//   - anchor: anchor point for cropping (Center, Top, Bottom, Left, Right, TopLeft, TopRight, BottomLeft, BottomRight)
//
// The anchor determines which part of the image to keep when cropping.
// For example, using TopLeft anchor will crop from the top-left corner.
//
// Available anchors:
//   - imaging.Center: crop from center
//   - imaging.Top: crop from top center
//   - imaging.Bottom: crop from bottom center
//   - imaging.Left: crop from left center
//   - imaging.Right: crop from right center
//   - imaging.TopLeft: crop from top-left corner
//   - imaging.TopRight: crop from top-right corner
//   - imaging.BottomLeft: crop from bottom-left corner
//   - imaging.BottomRight: crop from bottom-right corner
func CropAnchor(inputPath, outputPath string, width, height int, anchor imaging.Anchor) error {
	// Validate dimensions
	if width <= 0 {
		return fmt.Errorf("width must be positive, got %d", width)
	}
	if height <= 0 {
		return fmt.Errorf("height must be positive, got %d", height)
	}

	// Load the input image
	img, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", inputPath, err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop region fits within image
	if width > imgWidth {
		return fmt.Errorf("crop width %d exceeds image width %d", width, imgWidth)
	}
	if height > imgHeight {
		return fmt.Errorf("crop height %d exceeds image height %d", height, imgHeight)
	}

	// Use imaging.CropAnchor which handles anchor positioning
	cropped := imaging.CropAnchor(img, width, height, anchor)

	// Save the cropped image
	if err := imaging.Save(cropped, outputPath); err != nil {
		return fmt.Errorf("failed to save cropped image to %s: %w", outputPath, err)
	}

	return nil
}
