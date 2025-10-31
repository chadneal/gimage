package tools

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chadneal/gimage/internal/mcp"
)

// RegisterConvertImageTool registers the convert_image tool
func RegisterConvertImageTool(server *mcp.MCPServer) {
	tool := mcp.Tool{
		Name:        "convert_image",
		Description: "Convert an image between different formats (PNG, JPG/JPEG, WebP, GIF, TIFF, BMP). Useful for web optimization (converting to WebP), compatibility (PNG to JPG), or specific application requirements. Format detection is automatic based on file extension.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type":        "string",
					"description": "Input image file path (absolute or relative path)",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"png", "jpg", "jpeg", "webp", "gif", "tiff", "bmp"},
					"description": "Target image format. WebP recommended for web, PNG for lossless, JPG for photos.",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output file path. If not provided, generates filename with new extension (e.g., image.webp)",
				},
			},
			"required": []string{"input", "format"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			// Validate input
			input, err := validateString(args["input"], "input")
			if err != nil {
				return nil, err
			}

			// Validate format
			format, err := validateString(args["format"], "format")
			if err != nil {
				return nil, err
			}

			// Normalize format
			format = strings.ToLower(format)
			validFormats := map[string]bool{
				"png": true, "jpg": true, "jpeg": true,
				"webp": true, "gif": true, "tiff": true, "bmp": true,
			}
			if !validFormats[format] {
				return nil, fmt.Errorf("invalid format: %s (must be one of: png, jpg, jpeg, webp, gif, tiff, bmp)", format)
			}

			// Determine output path
			output, _ := args["output"].(string)
			if output == "" {
				// Change extension to new format
				ext := filepath.Ext(input)
				base := strings.TrimSuffix(input, ext)
				if format == "jpeg" {
					format = "jpg" // Use .jpg extension for JPEG
				}
				output = base + "." + format
			}

			// Get original format
			originalExt := filepath.Ext(input)
			originalFormat := strings.TrimPrefix(strings.ToLower(originalExt), ".")

			// Load image
			img, err := loadImage(input)
			if err != nil {
				return nil, fmt.Errorf("failed to load image: %w", err)
			}

			// Save in new format
			err = saveImage(img, output)
			if err != nil {
				return nil, fmt.Errorf("failed to save converted image: %w", err)
			}

			// Get file sizes
			originalSize, _ := getFileSize(input)
			newSize, _ := getFileSize(output)

			// Get absolute path for response
			absPath, _ := filepath.Abs(output)

			return map[string]interface{}{
				"success":         true,
				"output_path":     absPath,
				"original_format": originalFormat,
				"new_format":      format,
				"original_size":   formatBytes(originalSize),
				"new_size":        formatBytes(newSize),
			}, nil
		},
	}

	server.RegisterTool(tool)
}
