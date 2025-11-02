package tools

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/mcp"
)

func TestCropImageTool(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}

	// Create a test image (400x300)
	testImagePath := filepath.Join(tmpDir, "test.png")
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}
	file, err := os.Create(testImagePath)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	// Create server and register tool
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterCropImageTool(server)

	tests := []struct {
		name       string
		args       map[string]interface{}
		wantError  bool
		checkWidth int
	}{
		{
			name: "valid crop from top-left",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      0.0,
				"width":  200.0,
				"height": 150.0,
			},
			wantError:  false,
			checkWidth: 200,
		},
		{
			name: "valid crop from middle",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      100.0,
				"y":      75.0,
				"width":  200.0,
				"height": 150.0,
			},
			wantError:  false,
			checkWidth: 200,
		},
		{
			name: "crop with custom output path",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      0.0,
				"width":  100.0,
				"height": 100.0,
				"output": filepath.Join(tmpDir, "custom_crop.png"),
			},
			wantError:  false,
			checkWidth: 100,
		},
		{
			name: "small crop region",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      50.0,
				"y":      50.0,
				"width":  10.0,
				"height": 10.0,
			},
			wantError:  false,
			checkWidth: 10,
		},
		{
			name: "crop at bottom-right edge",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      200.0,
				"y":      150.0,
				"width":  200.0,
				"height": 150.0,
			},
			wantError:  false,
			checkWidth: 200,
		},
		{
			name: "missing x coordinate",
			args: map[string]interface{}{
				"input":  testImagePath,
				"y":      0.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "missing y coordinate",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "missing width",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      0.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "missing height",
			args: map[string]interface{}{
				"input": testImagePath,
				"x":     0.0,
				"y":     0.0,
				"width": 100.0,
			},
			wantError: true,
		},
		{
			name: "negative x coordinate",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      -10.0,
				"y":      0.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "negative y coordinate",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      -10.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "zero width",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      0.0,
				"width":  0.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "zero height",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      0.0,
				"width":  100.0,
				"height": 0.0,
			},
			wantError: true,
		},
		{
			name: "crop region exceeds width",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      300.0,
				"y":      0.0,
				"width":  200.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "crop region exceeds height",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      250.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "x coordinate outside image",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      500.0,
				"y":      0.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "y coordinate outside image",
			args: map[string]interface{}{
				"input":  testImagePath,
				"x":      0.0,
				"y":      400.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
		{
			name: "nonexistent input file",
			args: map[string]interface{}{
				"input":  filepath.Join(tmpDir, "nonexistent.png"),
				"x":      0.0,
				"y":      0.0,
				"width":  100.0,
				"height": 100.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := server.GetTool("crop_image")
			if tool == nil {
				t.Fatal("crop_image tool not registered")
			}

			result, err := tool.Handler(tt.args)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify result structure
			success, ok := result["success"].(bool)
			if !ok || !success {
				t.Error("Expected success: true in result")
			}

			outputPath, ok := result["output_path"].(string)
			if !ok {
				t.Error("Expected output_path in result")
				return
			}

			// Verify output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file not created: %s", outputPath)
				return
			}

			// Verify dimensions if checkWidth is set
			if tt.checkWidth > 0 {
				width, height, err := getImageDimensions(outputPath)
				if err != nil {
					t.Errorf("Failed to get dimensions: %v", err)
					return
				}

				if width != tt.checkWidth {
					t.Errorf("Expected width %d, got %d", tt.checkWidth, width)
				}

				expectedHeight := int(tt.args["height"].(float64))
				if height != expectedHeight {
					t.Errorf("Expected height %d, got %d", expectedHeight, height)
				}
			}

			// Verify crop_region in result
			cropRegion, ok := result["crop_region"].(string)
			if !ok {
				t.Error("Expected crop_region in result")
			} else {
				t.Logf("Crop region: %s", cropRegion)
			}

			// Verify crop_size in result
			cropSize, ok := result["crop_size"].(string)
			if !ok {
				t.Error("Expected crop_size in result")
			} else {
				t.Logf("Crop size: %s", cropSize)
			}
		})
	}
}

func TestCropImageToolSchema(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterCropImageTool(server)

	tool := server.GetTool("crop_image")
	if tool == nil {
		t.Fatal("crop_image tool not registered")
	}

	// Verify schema structure
	schema := tool.InputSchema
	if schema["type"] != "object" {
		t.Error("Expected type 'object' in schema")
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties in schema")
	}

	// Check required fields
	requiredFields := []string{"input", "x", "y", "width", "height"}
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required array in schema")
	}

	for _, field := range requiredFields {
		found := false
		for _, req := range required {
			if req == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required field '%s' not in schema", field)
		}

		if _, exists := properties[field]; !exists {
			t.Errorf("Property '%s' not defined in schema", field)
		}
	}

	// Verify optional output field exists
	if _, exists := properties["output"]; !exists {
		t.Error("Optional 'output' property not defined in schema")
	}

	// Verify coordinate fields have minimum values
	xProp, ok := properties["x"].(map[string]interface{})
	if !ok {
		t.Error("Expected x property to be an object")
	} else {
		if min, ok := xProp["minimum"].(int); !ok || min != 0 {
			t.Error("Expected x to have minimum value of 0")
		}
	}

	yProp, ok := properties["y"].(map[string]interface{})
	if !ok {
		t.Error("Expected y property to be an object")
	} else {
		if min, ok := yProp["minimum"].(int); !ok || min != 0 {
			t.Error("Expected y to have minimum value of 0")
		}
	}

	// Verify dimension fields have minimum values
	widthProp, ok := properties["width"].(map[string]interface{})
	if !ok {
		t.Error("Expected width property to be an object")
	} else {
		if min, ok := widthProp["minimum"].(int); !ok || min != 1 {
			t.Error("Expected width to have minimum value of 1")
		}
	}

	heightProp, ok := properties["height"].(map[string]interface{})
	if !ok {
		t.Error("Expected height property to be an object")
	} else {
		if min, ok := heightProp["minimum"].(int); !ok || min != 1 {
			t.Error("Expected height to have minimum value of 1")
		}
	}
}
