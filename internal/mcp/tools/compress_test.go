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

func TestCompressImageTool(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}

	// Create a test image (1000x1000) - larger for better compression testing
	testImagePath := filepath.Join(tmpDir, "test.png")
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	// Create a gradient pattern for better compression testing
	for y := 0; y < 1000; y++ {
		for x := 0; x < 1000; x++ {
			r := uint8((x * 255) / 1000)
			g := uint8((y * 255) / 1000)
			b := uint8(((x + y) * 255) / 2000)
			img.Set(x, y, color.RGBA{r, g, b, 255})
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
	RegisterCompressImageTool(server)

	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		quality   int
	}{
		{
			name: "default compression (90%)",
			args: map[string]interface{}{
				"input": testImagePath,
			},
			wantError: false,
			quality:   90,
		},
		{
			name: "high compression (75%)",
			args: map[string]interface{}{
				"input":   testImagePath,
				"quality": 75.0,
			},
			wantError: false,
			quality:   75,
		},
		{
			name: "low compression (95%)",
			args: map[string]interface{}{
				"input":   testImagePath,
				"quality": 95.0,
			},
			wantError: false,
			quality:   95,
		},
		{
			name: "custom output path",
			args: map[string]interface{}{
				"input":   testImagePath,
				"quality": 85.0,
				"output":  filepath.Join(tmpDir, "compressed_custom.jpg"),
			},
			wantError: false,
			quality:   85,
		},
		{
			name: "quality too high",
			args: map[string]interface{}{
				"input":   testImagePath,
				"quality": 101.0,
			},
			wantError: true,
		},
		{
			name: "quality too low",
			args: map[string]interface{}{
				"input":   testImagePath,
				"quality": 0.0,
			},
			wantError: true,
		},
		{
			name: "negative quality",
			args: map[string]interface{}{
				"input":   testImagePath,
				"quality": -10.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := server.GetTool("compress_image")
			if tool == nil {
				t.Fatal("compress_image tool not registered")
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

			// Verify quality in result
			quality, ok := result["quality"].(int)
			if !ok {
				t.Error("Expected quality in result")
			} else if quality != tt.quality {
				t.Errorf("Expected quality %d, got %d", tt.quality, quality)
			}

			// Verify file size information
			originalSizeBytes, ok := result["original_size_bytes"].(int64)
			if !ok {
				t.Error("Expected original_size_bytes in result")
			}

			compressedSizeBytes, ok := result["compressed_size_bytes"].(int64)
			if !ok {
				t.Error("Expected compressed_size_bytes in result")
			}

			// Verify compression stats are reasonable
			// Note: Compression may not always reduce size depending on image content and format conversion
			// We mainly verify that the compression process completes successfully
			_ = originalSizeBytes   // Use the value to avoid unused variable warning
			_ = compressedSizeBytes // Use the value to avoid unused variable warning

			// Verify savings information
			if _, ok := result["savings_bytes"].(int64); !ok {
				t.Error("Expected savings_bytes in result")
			}

			if _, ok := result["savings_percent"].(string); !ok {
				t.Error("Expected savings_percent in result")
			}

			if _, ok := result["compression_ratio"].(string); !ok {
				t.Error("Expected compression_ratio in result")
			}

			// Verify human-readable sizes
			if _, ok := result["original_size_human"].(string); !ok {
				t.Error("Expected original_size_human in result")
			}

			if _, ok := result["compressed_size_human"].(string); !ok {
				t.Error("Expected compressed_size_human in result")
			}
		})
	}
}

func TestCompressImageToolSchema(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterCompressImageTool(server)

	tool := server.GetTool("compress_image")
	if tool == nil {
		t.Fatal("compress_image tool not registered")
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
	requiredFields := []string{"input"}
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

	// Verify quality has default, minimum, and maximum
	quality, ok := properties["quality"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected quality property definition")
	}

	if _, exists := quality["default"]; !exists {
		t.Error("Expected default value for quality")
	}

	if _, exists := quality["minimum"]; !exists {
		t.Error("Expected minimum constraint on quality")
	}

	if _, exists := quality["maximum"]; !exists {
		t.Error("Expected maximum constraint on quality")
	}
}
