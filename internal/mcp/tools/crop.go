package tools

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/apresai/gimage/internal/mcp"
	"github.com/disintegration/imaging"
)

// RegisterCropImageTool registers the crop_image tool
func RegisterCropImageTool(server *mcp.MCPServer) {
	tool := mcp.Tool{
		Name:        "crop_image",
		Description: "Crop an image to a specific rectangular region. Specify the top-left corner coordinates (x, y) and the width and height of the region to extract. Useful for removing unwanted borders, focusing on specific areas, or extracting thumbnails from larger images.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type":        "string",
					"description": "Input image file path (absolute or relative path)",
				},
				"x": map[string]interface{}{
					"type":        "integer",
					"description": "X coordinate of top-left corner of crop region (0 = left edge)",
					"minimum":     0,
				},
				"y": map[string]interface{}{
					"type":        "integer",
					"description": "Y coordinate of top-left corner of crop region (0 = top edge)",
					"minimum":     0,
				},
				"width": map[string]interface{}{
					"type":        "integer",
					"description": "Width of crop region in pixels (must be positive)",
					"minimum":     1,
				},
				"height": map[string]interface{}{
					"type":        "integer",
					"description": "Height of crop region in pixels (must be positive)",
					"minimum":     1,
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output file path. If not provided, generates filename like input_cropped.ext",
				},
			},
			"required": []string{"input", "x", "y", "width", "height"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			// Validate input
			inputArg, err := validateString(args["input"], "input")
			if err != nil {
				return nil, err
			}
			input, err := ValidateInputPath(inputArg)
			if err != nil {
				return nil, fmt.Errorf("input validation failed: %w", err)
			}

			// Validate coordinates and dimensions
			x, err := validatePositiveInt(args["x"], "x")
			if err != nil {
				// x can be 0, so check differently
				xVal, ok := args["x"].(float64)
				if !ok {
					return nil, fmt.Errorf("x must be a number")
				}
				if xVal < 0 {
					return nil, fmt.Errorf("x must be non-negative")
				}
				x = int(xVal)
			}

			y, err := validatePositiveInt(args["y"], "y")
			if err != nil {
				// y can be 0, so check differently
				yVal, ok := args["y"].(float64)
				if !ok {
					return nil, fmt.Errorf("y must be a number")
				}
				if yVal < 0 {
					return nil, fmt.Errorf("y must be non-negative")
				}
				y = int(yVal)
			}

			width, err := validatePositiveInt(args["width"], "width")
			if err != nil {
				return nil, err
			}

			height, err := validatePositiveInt(args["height"], "height")
			if err != nil {
				return nil, err
			}

			// Determine output path
			outputArg, _ := args["output"].(string)
			defaultFilename := generateOutputPath(input, "cropped")
			pathResult, pathErr := ValidateAndFixOutputPath(outputArg, defaultFilename)
			if pathErr != nil {
				return nil, fmt.Errorf("output path validation failed: %w", pathErr)
			}
			output := pathResult.Path

			// Load image
			img, err := loadImage(input)
			if err != nil {
				return nil, fmt.Errorf("failed to load image: %w", err)
			}

			// Get image bounds
			bounds := img.Bounds()

			// Validate crop region is within image bounds
			if x < 0 || y < 0 {
				return nil, fmt.Errorf("x and y coordinates must be non-negative")
			}
			if x+width > bounds.Dx() || y+height > bounds.Dy() {
				return nil, fmt.Errorf("crop region (%d,%d,%d,%d) extends beyond image bounds (%dx%d)",
					x, y, width, height, bounds.Dx(), bounds.Dy())
			}

			// Create crop rectangle
			cropRect := image.Rect(x, y, x+width, y+height)

			// Crop image
			cropped := imaging.Crop(img, cropRect)

			// Save cropped image
			err = saveImage(cropped, output)
			if err != nil {
				return nil, fmt.Errorf("failed to save cropped image: %w", err)
			}

			// Get absolute path for response
			absPath, _ := filepath.Abs(output)

			result := map[string]interface{}{
				"success":     true,
				"output_path": absPath,
				"crop_region": fmt.Sprintf("(%d,%d,%d,%d)", x, y, width, height),
				"crop_size":   fmt.Sprintf("%dx%d", width, height),
			}
			if pathResult.Warning != "" {
				result["warning"] = pathResult.Warning
			}
			return result, nil
		},
	}

	server.RegisterTool(tool)
}
