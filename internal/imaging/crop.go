// Package imaging provides image processing operations using pure Go.
package imaging

import (
	"context"
	"fmt"
	"image"

	"github.com/apresai/gimage/internal/progress"
	"github.com/disintegration/imaging"
)

// CropImage crops a rectangular region from an image.
//
// Parameters:
//   - ctx: context for cancellation support
//   - inputPath: path to input image
//   - outputPath: path to save cropped image
//   - x, y: top-left corner coordinates of the crop region
//   - width, height: dimensions of crop region
//
// Progress reporting can be provided via context using progress.WithReporter.
//
// Returns error if:
//   - context is cancelled
//   - input file does not exist or cannot be read
//   - crop region is outside image bounds
//   - dimensions are not positive
//   - output cannot be written
func CropImage(ctx context.Context, inputPath, outputPath string, x, y, width, height int) error {
	reporter := progress.FromContext(ctx)
	reporter.Start(ctx, fmt.Sprintf("Cropping image region %dx%d at (%d,%d)", width, height, x, y))

	// Validate dimensions
	if width <= 0 {
		err := fmt.Errorf("width must be positive, got %d", width)
		reporter.Error(err)
		return err
	}
	if height <= 0 {
		err := fmt.Errorf("height must be positive, got %d", height)
		reporter.Error(err)
		return err
	}

	// Validate coordinates
	if x < 0 {
		err := fmt.Errorf("x coordinate must be non-negative, got %d", x)
		reporter.Error(err)
		return err
	}
	if y < 0 {
		err := fmt.Errorf("y coordinate must be non-negative, got %d", y)
		reporter.Error(err)
		return err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Load the input image
	reporter.Update(1, 4, "Loading input image")
	img, err := imaging.Open(inputPath)
	if err != nil {
		err = fmt.Errorf("failed to open image %s: %w", inputPath, err)
		reporter.Error(err)
		return err
	}

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop region is within image bounds
	reporter.Update(2, 4, "Validating crop region")
	if x >= imgWidth {
		err := fmt.Errorf("x coordinate %d is outside image width %d", x, imgWidth)
		reporter.Error(err)
		return err
	}
	if y >= imgHeight {
		err := fmt.Errorf("y coordinate %d is outside image height %d", y, imgHeight)
		reporter.Error(err)
		return err
	}
	if x+width > imgWidth {
		err := fmt.Errorf("crop region (x=%d + width=%d = %d) exceeds image width %d", x, width, x+width, imgWidth)
		reporter.Error(err)
		return err
	}
	if y+height > imgHeight {
		err := fmt.Errorf("crop region (y=%d + height=%d = %d) exceeds image height %d", y, height, y+height, imgHeight)
		reporter.Error(err)
		return err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Create crop rectangle
	cropRect := image.Rect(x, y, x+width, y+height)

	// Perform the crop
	reporter.Update(3, 4, "Cropping image")
	cropped := imaging.Crop(img, cropRect)

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Save the cropped image
	reporter.Update(4, 4, "Saving cropped image")
	if err := imaging.Save(cropped, outputPath); err != nil {
		err = fmt.Errorf("failed to save cropped image to %s: %w", outputPath, err)
		reporter.Error(err)
		return err
	}

	reporter.Complete(outputPath)
	return nil
}

// CropCenter crops a region from the center of the image.
//
// Parameters:
//   - ctx: context for cancellation support
//   - inputPath: path to input image
//   - outputPath: path to save cropped image
//   - width, height: dimensions of crop region
//
// The crop region will be centered on the image. If the requested dimensions
// are larger than the image, an error is returned.
//
// Progress reporting can be provided via context using progress.WithReporter.
func CropCenter(ctx context.Context, inputPath, outputPath string, width, height int) error {
	reporter := progress.FromContext(ctx)
	reporter.Start(ctx, fmt.Sprintf("Cropping %dx%d from center", width, height))

	// Validate dimensions
	if width <= 0 {
		err := fmt.Errorf("width must be positive, got %d", width)
		reporter.Error(err)
		return err
	}
	if height <= 0 {
		err := fmt.Errorf("height must be positive, got %d", height)
		reporter.Error(err)
		return err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Load the input image
	reporter.Update(1, 3, "Loading input image")
	img, err := imaging.Open(inputPath)
	if err != nil {
		err = fmt.Errorf("failed to open image %s: %w", inputPath, err)
		reporter.Error(err)
		return err
	}

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop region fits within image
	if width > imgWidth {
		err := fmt.Errorf("crop width %d exceeds image width %d", width, imgWidth)
		reporter.Error(err)
		return err
	}
	if height > imgHeight {
		err := fmt.Errorf("crop height %d exceeds image height %d", height, imgHeight)
		reporter.Error(err)
		return err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Use imaging.CropCenter which handles center calculation
	reporter.Update(2, 3, "Cropping from center")
	cropped := imaging.CropCenter(img, width, height)

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Save the cropped image
	reporter.Update(3, 3, "Saving cropped image")
	if err := imaging.Save(cropped, outputPath); err != nil {
		err = fmt.Errorf("failed to save cropped image to %s: %w", outputPath, err)
		reporter.Error(err)
		return err
	}

	reporter.Complete(outputPath)
	return nil
}

// CropAnchor crops a region with a specific anchor point.
//
// Parameters:
//   - ctx: context for cancellation support
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
//
// Progress reporting can be provided via context using progress.WithReporter.
func CropAnchor(ctx context.Context, inputPath, outputPath string, width, height int, anchor imaging.Anchor) error {
	reporter := progress.FromContext(ctx)
	reporter.Start(ctx, fmt.Sprintf("Cropping %dx%d with anchor", width, height))

	// Validate dimensions
	if width <= 0 {
		err := fmt.Errorf("width must be positive, got %d", width)
		reporter.Error(err)
		return err
	}
	if height <= 0 {
		err := fmt.Errorf("height must be positive, got %d", height)
		reporter.Error(err)
		return err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Load the input image
	reporter.Update(1, 3, "Loading input image")
	img, err := imaging.Open(inputPath)
	if err != nil {
		err = fmt.Errorf("failed to open image %s: %w", inputPath, err)
		reporter.Error(err)
		return err
	}

	// Get image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop region fits within image
	if width > imgWidth {
		err := fmt.Errorf("crop width %d exceeds image width %d", width, imgWidth)
		reporter.Error(err)
		return err
	}
	if height > imgHeight {
		err := fmt.Errorf("crop height %d exceeds image height %d", height, imgHeight)
		reporter.Error(err)
		return err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Use imaging.CropAnchor which handles anchor positioning
	reporter.Update(2, 3, "Cropping with anchor")
	cropped := imaging.CropAnchor(img, width, height, anchor)

	// Check for cancellation
	select {
	case <-ctx.Done():
		err := fmt.Errorf("operation cancelled: %w", ctx.Err())
		reporter.Error(err)
		return err
	default:
	}

	// Save the cropped image
	reporter.Update(3, 3, "Saving cropped image")
	if err := imaging.Save(cropped, outputPath); err != nil {
		err = fmt.Errorf("failed to save cropped image to %s: %w", outputPath, err)
		reporter.Error(err)
		return err
	}

	reporter.Complete(outputPath)
	return nil
}
