package tools

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	gimaging "github.com/apresai/gimage/internal/imaging"
)

// generateOutputPath creates an output path based on input and suffix
func generateOutputPath(input, suffix string) string {
	ext := filepath.Ext(input)
	base := strings.TrimSuffix(filepath.Base(input), ext)
	return fmt.Sprintf("%s_%s%s", base, suffix, ext)
}

// generateTimestampedPath creates a path with timestamp
func generateTimestampedPath(prefix, ext string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d.%s", prefix, timestamp, ext)
}

// getImageDimensions returns the width and height of an image
func getImageDimensions(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}

	return img.Width, img.Height, nil
}

// getFileSize returns the size of a file in bytes
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	return info.Size(), nil
}

// validatePositiveInt validates that a value is a positive integer
func validatePositiveInt(value interface{}, name string) (int, error) {
	num, ok := value.(float64) // JSON numbers are float64
	if !ok {
		return 0, fmt.Errorf("%s must be a number", name)
	}
	if num < 1 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return int(num), nil
}

// validateString validates that a value is a non-empty string
func validateString(value interface{}, name string) (string, error) {
	str, ok := value.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return str, nil
}

// loadImage loads an image from a file
func loadImage(path string) (image.Image, error) {
	return imaging.Open(path)
}

// saveImage saves an image to a file
func saveImage(img image.Image, path string) error {
	// Use internal/imaging package which now supports WebP via nativewebp
	format := filepath.Ext(path)
	format = strings.ToLower(strings.TrimPrefix(format, "."))

	// For WebP, use our custom implementation
	if format == "webp" {
		return gimaging.SaveImageWithFormat(img, path, format)
	}

	// For other formats, use imaging.Save
	return imaging.Save(img, path)
}

// isVertexModel checks if a model is a Vertex AI model
func isVertexModel(model string) bool {
	vertexModels := []string{
		"imagen-3.0-generate-002",
		"imagen-4",
		"imagen-4-standard",
		"imagen-4-ultra",
		"imagen-4-fast",
	}
	for _, vm := range vertexModels {
		if model == vm {
			return true
		}
	}
	return false
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
