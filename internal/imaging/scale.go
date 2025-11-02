// Package imaging provides image processing operations using pure Go.
package imaging

import (
	"fmt"

	"github.com/disintegration/imaging"
)

// ScaleImage scales an image by a factor using high-quality Lanczos resampling.
//
// Parameters:
//   - inputPath: path to input image
//   - outputPath: path to save scaled image
//   - factor: scaling factor (e.g., 0.5 for half size, 2.0 for double size)
//
// The image dimensions will be multiplied by the factor. For example:
//   - factor 0.5 = 50% size (half)
//   - factor 1.0 = same size (no change)
//   - factor 2.0 = 200% size (double)
//
// Returns error if:
//   - input file does not exist or cannot be read
//   - factor is not positive
//   - output cannot be written
func ScaleImage(inputPath, outputPath string, factor float64) error {
	// Validate factor
	if factor <= 0 {
		return fmt.Errorf("factor must be positive, got %f", factor)
	}

	// Load the input image
	img, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", inputPath, err)
	}

	// Get current dimensions
	bounds := img.Bounds()
	currentWidth := bounds.Dx()
	currentHeight := bounds.Dy()

	// Calculate new dimensions
	newWidth := int(float64(currentWidth) * factor)
	newHeight := int(float64(currentHeight) * factor)

	// Ensure dimensions are at least 1
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	// Scale using Lanczos resampling (high quality)
	scaled := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

	// Save the scaled image
	if err := imaging.Save(scaled, outputPath); err != nil {
		return fmt.Errorf("failed to save scaled image to %s: %w", outputPath, err)
	}

	return nil
}
