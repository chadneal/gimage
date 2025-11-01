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

func TestScaleImageTool(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}

	// Create a test image (200x200)
	testImagePath := filepath.Join(tmpDir, "test.png")
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
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
	RegisterScaleImageTool(server)

	tests := []struct {
		name          string
		args          map[string]interface{}
		wantError     bool
		expectedWidth int
	}{
		{
			name: "scale to half size",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": 0.5,
			},
			wantError:     false,
			expectedWidth: 100,
		},
		{
			name: "scale to double size",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": 2.0,
			},
			wantError:     false,
			expectedWidth: 400,
		},
		{
			name: "scale with custom output",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": 0.75,
				"output": filepath.Join(tmpDir, "scaled_custom.png"),
			},
			wantError:     false,
			expectedWidth: 150,
		},
		{
			name: "missing factor",
			args: map[string]interface{}{
				"input": testImagePath,
			},
			wantError: true,
		},
		{
			name: "factor too small",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": 0.05,
			},
			wantError: true,
		},
		{
			name: "factor too large",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": 15.0,
			},
			wantError: true,
		},
		{
			name: "zero factor",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": 0.0,
			},
			wantError: true,
		},
		{
			name: "negative factor",
			args: map[string]interface{}{
				"input":  testImagePath,
				"factor": -0.5,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := server.GetTool("scale_image")
			if tool == nil {
				t.Fatal("scale_image tool not registered")
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

			// Verify dimensions (aspect ratio should be maintained)
			width, height, err := getImageDimensions(outputPath)
			if err != nil {
				t.Errorf("Failed to get dimensions: %v", err)
				return
			}

			if width != tt.expectedWidth {
				t.Errorf("Expected width %d, got %d", tt.expectedWidth, width)
			}

			// Verify aspect ratio is maintained (should be square)
			if width != height {
				t.Errorf("Aspect ratio not maintained: %dx%d", width, height)
			}

			// Verify scale_factor in result
			scaleFactor, ok := result["scale_factor"].(float64)
			if !ok {
				t.Error("Expected scale_factor in result")
			} else {
				expectedFactor := tt.args["factor"].(float64)
				if scaleFactor != expectedFactor {
					t.Errorf("Expected scale_factor %.2f, got %.2f", expectedFactor, scaleFactor)
				}
			}
		})
	}
}

func TestScaleImageToolSchema(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterScaleImageTool(server)

	tool := server.GetTool("scale_image")
	if tool == nil {
		t.Fatal("scale_image tool not registered")
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
	requiredFields := []string{"input", "factor"}
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

	// Verify factor has minimum and maximum
	factor, ok := properties["factor"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected factor property definition")
	}

	if _, exists := factor["minimum"]; !exists {
		t.Error("Expected minimum constraint on factor")
	}

	if _, exists := factor["maximum"]; !exists {
		t.Error("Expected maximum constraint on factor")
	}
}
