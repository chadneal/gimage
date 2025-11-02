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

func TestConvertImageTool(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}

	// Create a test PNG image
	testImagePath := filepath.Join(tmpDir, "test.png")
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
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
	RegisterConvertImageTool(server)

	tests := []struct {
		name         string
		args         map[string]interface{}
		wantError    bool
		expectFormat string
	}{
		{
			name: "PNG to JPG",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "jpg",
			},
			wantError:    false,
			expectFormat: "jpg",
		},
		{
			name: "PNG to WebP",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "webp",
			},
			wantError:    false,
			expectFormat: "webp",
		},
		{
			name: "PNG to GIF",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "gif",
			},
			wantError:    false,
			expectFormat: "gif",
		},
		{
			name: "PNG to TIFF",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "tiff",
			},
			wantError:    false,
			expectFormat: "tiff",
		},
		{
			name: "PNG to BMP",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "bmp",
			},
			wantError:    false,
			expectFormat: "bmp",
		},
		{
			name: "convert with custom output path",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "jpg",
				"output": filepath.Join(tmpDir, "custom_output.jpg"),
			},
			wantError:    false,
			expectFormat: "jpg",
		},
		{
			name: "JPEG format variant",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "jpeg",
			},
			wantError:    false,
			expectFormat: "jpeg",
		},
		{
			name: "uppercase format",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "JPG",
			},
			wantError:    false,
			expectFormat: "jpg",
		},
		{
			name: "missing input",
			args: map[string]interface{}{
				"format": "jpg",
			},
			wantError: true,
		},
		{
			name: "missing format",
			args: map[string]interface{}{
				"input": testImagePath,
			},
			wantError: true,
		},
		{
			name: "invalid format",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "invalid",
			},
			wantError: true,
		},
		{
			name: "nonexistent input file",
			args: map[string]interface{}{
				"input":  filepath.Join(tmpDir, "nonexistent.png"),
				"format": "jpg",
			},
			wantError: true,
		},
		{
			name: "empty format",
			args: map[string]interface{}{
				"input":  testImagePath,
				"format": "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := server.GetTool("convert_image")
			if tool == nil {
				t.Fatal("convert_image tool not registered")
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

			// Verify format in result
			newFormat, ok := result["new_format"].(string)
			if !ok {
				t.Error("Expected new_format in result")
			} else if newFormat != tt.expectFormat {
				t.Errorf("Expected format %s, got %s", tt.expectFormat, newFormat)
			}

			// Verify original format
			originalFormat, ok := result["original_format"].(string)
			if !ok {
				t.Error("Expected original_format in result")
			} else if originalFormat != "png" {
				t.Errorf("Expected original format png, got %s", originalFormat)
			}

			// Verify sizes are present
			if _, ok := result["original_size"]; !ok {
				t.Error("Expected original_size in result")
			}
			if _, ok := result["new_size"]; !ok {
				t.Error("Expected new_size in result")
			}
		})
	}
}

func TestConvertImageToolSchema(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterConvertImageTool(server)

	tool := server.GetTool("convert_image")
	if tool == nil {
		t.Fatal("convert_image tool not registered")
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
	requiredFields := []string{"input", "format"}
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

	// Verify format enum
	formatProp, ok := properties["format"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected format property to be an object")
	}

	enumValues, ok := formatProp["enum"].([]string)
	if !ok {
		t.Error("Expected format to have enum values")
	} else {
		expectedFormats := []string{"png", "jpg", "jpeg", "webp", "gif", "tiff", "bmp"}
		if len(enumValues) != len(expectedFormats) {
			t.Errorf("Expected %d format options, got %d", len(expectedFormats), len(enumValues))
		}
	}
}
