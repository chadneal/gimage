// Package imaging provides image processing operations using pure Go.
package imaging

import (
	"fmt"

	"github.com/disintegration/imaging"
)

// ResizeImage resizes an image to specific dimensions using high-quality Lanczos resampling.
//
// Parameters:
//   - inputPath: path to input image
//   - outputPath: path to save resized image
//   - width: target width in pixels
//   - height: target height in pixels
//
// The image will be resized to exactly the specified dimensions. If you want to preserve
// aspect ratio, use ResizeFit or calculate one dimension based on the other.
//
// Returns error if:
//   - input file does not exist or cannot be read
//   - dimensions are not positive
//   - output cannot be written
func ResizeImage(inputPath, outputPath string, width, height int) error {
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

	// Resize using Lanczos resampling (high quality)
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// Save the resized image
	if err := imaging.Save(resized, outputPath); err != nil {
		return fmt.Errorf("failed to save resized image to %s: %w", outputPath, err)
	}

	return nil
}

// ResizeFit resizes an image to fit within specified dimensions while preserving aspect ratio.
//
// Parameters:
//   - inputPath: path to input image
//   - outputPath: path to save resized image
//   - width: maximum width in pixels
//   - height: maximum height in pixels
//
// The image will be resized to fit within the specified dimensions while maintaining
// its original aspect ratio. The resulting image may be smaller than the specified
// dimensions in one or both directions.
func ResizeFit(inputPath, outputPath string, width, height int) error {
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

	// Resize to fit within bounds, preserving aspect ratio
	resized := imaging.Fit(img, width, height, imaging.Lanczos)

	// Save the resized image
	if err := imaging.Save(resized, outputPath); err != nil {
		return fmt.Errorf("failed to save resized image to %s: %w", outputPath, err)
	}

	return nil
}
