package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chadneal/gimage/internal/imaging"
	"github.com/chadneal/gimage/pkg/models"
)

const (
	defaultOutputDir    = "."
	defaultOutputPrefix = "generated"
	defaultFilePerms    = 0644
	defaultDirPerms     = 0755
)

// SaveImage saves a generated image to disk at the specified path.
// If the directory doesn't exist, it will be created.
// If the output path has a different extension than the source format,
// the image will be automatically converted to the target format.
func SaveImage(img *models.GeneratedImage, outputPath string) error {
	if img == nil {
		return fmt.Errorf("image cannot be nil")
	}

	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	if len(img.Data) == 0 {
		return fmt.Errorf("image data is empty")
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, defaultDirPerms); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Get target format from output path
	targetFormat := imaging.ExtractFormatFromPath(outputPath)
	sourceFormat := normalizeFormat(img.Format)

	// If formats don't match, convert the image
	if targetFormat != sourceFormat {
		convertedData, err := imaging.ConvertImageData(img.Data, targetFormat)
		if err != nil {
			return fmt.Errorf("failed to convert image from %s to %s: %w", sourceFormat, targetFormat, err)
		}

		// Write the converted image data
		if err := os.WriteFile(outputPath, convertedData, defaultFilePerms); err != nil {
			return fmt.Errorf("failed to write image to %s: %w", outputPath, err)
		}

		return nil
	}

	// Write the image data as-is (no conversion needed)
	if err := os.WriteFile(outputPath, img.Data, defaultFilePerms); err != nil {
		return fmt.Errorf("failed to write image to %s: %w", outputPath, err)
	}

	return nil
}

// SaveImageWithMetadata saves an image and creates a companion metadata file
func SaveImageWithMetadata(img *models.GeneratedImage, outputPath string) error {
	// Save the image
	if err := SaveImage(img, outputPath); err != nil {
		return err
	}

	// Save metadata in a .json file alongside the image
	metadataPath := outputPath + ".json"
	if err := saveMetadata(img, metadataPath); err != nil {
		// Log error but don't fail the save operation
		return fmt.Errorf("warning: failed to save metadata: %w", err)
	}

	return nil
}

// saveMetadata writes image metadata to a JSON file
func saveMetadata(img *models.GeneratedImage, path string) error {
	// Build JSON manually to avoid importing encoding/json
	var metadata string
	metadata = "{\n"

	first := true
	for key, value := range img.Metadata {
		if !first {
			metadata += ",\n"
		}
		metadata += fmt.Sprintf("  %q: %q", key, value)
		first = false
	}

	metadata += fmt.Sprintf(",\n  \"format\": %q", img.Format)
	metadata += fmt.Sprintf(",\n  \"width\": %d", img.Width)
	metadata += fmt.Sprintf(",\n  \"height\": %d", img.Height)
	metadata += fmt.Sprintf(",\n  \"size_bytes\": %d", len(img.Data))
	metadata += "\n}\n"

	return os.WriteFile(path, []byte(metadata), defaultFilePerms)
}

// GenerateOutputPath generates a default output path with timestamp
// Format: generated_YYYYMMDD_HHMMSS.{format}
func GenerateOutputPath(format string) string {
	// Use current time for filename
	timestamp := time.Now().Format("20060102_150405")

	// Ensure format has no leading dot
	format = normalizeFormat(format)

	filename := fmt.Sprintf("%s_%s.%s", defaultOutputPrefix, timestamp, format)

	return filepath.Join(defaultOutputDir, filename)
}

// GenerateOutputPathWithPrefix generates an output path with a custom prefix
func GenerateOutputPathWithPrefix(prefix, format string) string {
	if prefix == "" {
		prefix = defaultOutputPrefix
	}

	timestamp := time.Now().Format("20060102_150405")
	format = normalizeFormat(format)

	filename := fmt.Sprintf("%s_%s.%s", prefix, timestamp, format)

	return filepath.Join(defaultOutputDir, filename)
}

// GenerateOutputPathInDir generates an output path in a specific directory
func GenerateOutputPathInDir(dir, format string) string {
	if dir == "" {
		dir = defaultOutputDir
	}

	timestamp := time.Now().Format("20060102_150405")
	format = normalizeFormat(format)

	filename := fmt.Sprintf("%s_%s.%s", defaultOutputPrefix, timestamp, format)

	return filepath.Join(dir, filename)
}

// GenerateOutputPathCustom generates an output path with full customization
func GenerateOutputPathCustom(dir, prefix, format string) string {
	if dir == "" {
		dir = defaultOutputDir
	}
	if prefix == "" {
		prefix = defaultOutputPrefix
	}

	timestamp := time.Now().Format("20060102_150405")
	format = normalizeFormat(format)

	filename := fmt.Sprintf("%s_%s.%s", prefix, timestamp, format)

	return filepath.Join(dir, filename)
}

// normalizeFormat removes leading dots and converts to lowercase
// Also normalizes format variations (jpg/jpeg, tif/tiff) for comparison
func normalizeFormat(format string) string {
	if format == "" {
		return "png"
	}

	// Remove leading dot if present
	if format[0] == '.' {
		format = format[1:]
	}

	// Convert to lowercase
	format = toLower(format)

	// Normalize variations
	switch format {
	case "jpg":
		return "jpeg"
	case "tif":
		return "tiff"
	default:
		return format
	}
}

// toLower converts a string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// ValidateOutputPath checks if an output path is valid and writable
func ValidateOutputPath(path string) error {
	if path == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	// Check if parent directory exists or can be created
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				// Try to create the directory
				if err := os.MkdirAll(dir, defaultDirPerms); err != nil {
					return fmt.Errorf("cannot create directory %s: %w", dir, err)
				}
			} else {
				return fmt.Errorf("cannot access directory %s: %w", dir, err)
			}
		} else if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dir)
		}
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	return nil
}

// EnsureOutputDir ensures the output directory exists
func EnsureOutputDir(dir string) error {
	if dir == "" || dir == "." || dir == "/" {
		return nil
	}

	if err := os.MkdirAll(dir, defaultDirPerms); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return nil
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GenerateUniqueOutputPath generates a unique output path, incrementing if file exists
func GenerateUniqueOutputPath(format string) string {
	basePath := GenerateOutputPath(format)

	// If file doesn't exist, return as-is
	if !FileExists(basePath) {
		return basePath
	}

	// Extract components
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(basePath)
	nameWithoutExt := basePath[:len(basePath)-len(ext)]

	// Try adding numbers until we find a unique name
	for i := 1; i < 1000; i++ {
		newPath := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		if !FileExists(newPath) {
			return newPath
		}
	}

	// Fallback: add timestamp with milliseconds
	timestamp := time.Now().Format("20060102_150405.000")
	format = normalizeFormat(format)
	filename := fmt.Sprintf("%s_%s.%s", defaultOutputPrefix, timestamp, format)

	return filepath.Join(dir, filename)
}
