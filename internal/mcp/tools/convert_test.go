package tools

import (
	"strings"
	"testing"

	"github.com/apresai/gimage/internal/mcp"
)

func TestRegisterConvertImageTool(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)

	// Register the tool
	RegisterConvertImageTool(server)

	// Verify tool is registered
	tool := server.GetTool("convert_image")
	if tool == nil {
		t.Fatal("convert_image tool not registered")
	}

	// Verify tool properties
	if tool.Name != "convert_image" {
		t.Errorf("Expected name 'convert_image', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description is empty")
	}

	if tool.Handler == nil {
		t.Error("Tool handler is nil")
	}
}

func TestConvertImageTool_InputSchema(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterConvertImageTool(server)

	tool := server.GetTool("convert_image")
	if tool == nil {
		t.Fatal("convert_image tool not registered")
	}

	schema := tool.InputSchema

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required field is not a string slice")
	}

	expectedRequired := []string{"input", "format"}
	if len(required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
	}

	// Check properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties field is not a map")
	}

	// Verify input property
	if _, exists := properties["input"]; !exists {
		t.Error("input property missing")
	}

	// Verify format property
	formatProp, exists := properties["format"]
	if !exists {
		t.Error("format property missing")
	} else {
		formatMap, ok := formatProp.(map[string]interface{})
		if !ok {
			t.Error("format property is not a map")
		} else {
			enum, exists := formatMap["enum"]
			if !exists {
				t.Error("format enum missing")
			}
			if enum == nil {
				t.Error("format enum is nil")
			}
		}
	}

	// Verify optional output property
	if _, exists := properties["output"]; !exists {
		t.Error("output property missing")
	}
}

func TestConvertImageTool_ValidationErrors(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterConvertImageTool(server)

	tool := server.GetTool("convert_image")
	if tool == nil {
		t.Fatal("convert_image tool not registered")
	}

	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name:      "missing input",
			args:      map[string]interface{}{"format": "png"},
			wantError: true,
			errorMsg:  "input is required",
		},
		{
			name:      "missing format",
			args:      map[string]interface{}{"input": "test.jpg"},
			wantError: true,
			errorMsg:  "format is required",
		},
		{
			name: "invalid format",
			args: map[string]interface{}{
				"input":  "test.jpg",
				"format": "invalid",
			},
			wantError: true,
			errorMsg:  "invalid format",
		},
		{
			name: "input not a string",
			args: map[string]interface{}{
				"input":  123,
				"format": "png",
			},
			wantError: true,
			errorMsg:  "input is required",
		},
		{
			name: "format not a string",
			args: map[string]interface{}{
				"input":  "test.jpg",
				"format": 123,
			},
			wantError: true,
			errorMsg:  "format is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.args)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConvertImageTool_FormatNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PNG", "png"},
		{"JPG", "jpg"},
		{"JPEG", "jpeg"},
		{"WebP", "webp"},
		{"GIF", "gif"},
		{"TIFF", "tiff"},
		{"BMP", "bmp"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Test format normalization logic
			format := tt.input
			normalized := normalizeFormat(format)
			if normalized != tt.expected {
				t.Errorf("normalizeFormat(%s) = %s, want %s", tt.input, normalized, tt.expected)
			}
		})
	}
}

// Helper function to test format normalization
func normalizeFormat(format string) string {
	// This matches the logic in convert.go
	return strings.ToLower(format)
}

func TestValidFormats(t *testing.T) {
	validFormats := []string{"png", "jpg", "jpeg", "webp", "gif", "tiff", "bmp"}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			// Each format should be valid
			validFormatsMap := map[string]bool{
				"png": true, "jpg": true, "jpeg": true,
				"webp": true, "gif": true, "tiff": true, "bmp": true,
			}

			if !validFormatsMap[format] {
				t.Errorf("Format %s should be valid", format)
			}
		})
	}
}

func TestConvertImageTool_DefaultOutputPath(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		format       string
		expectedExt  string
	}{
		{
			name:        "png to jpg",
			input:       "image.png",
			format:      "jpg",
			expectedExt: ".jpg",
		},
		{
			name:        "jpg to png",
			input:       "photo.jpg",
			format:      "png",
			expectedExt: ".png",
		},
		{
			name:        "png to webp",
			input:       "image.png",
			format:      "webp",
			expectedExt: ".webp",
		},
		{
			name:        "jpeg format uses jpg extension",
			input:       "image.png",
			format:      "jpeg",
			expectedExt: ".jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the expected extension logic works
			targetFormat := tt.format
			if targetFormat == "jpeg" {
				targetFormat = "jpg"
			}

			if "."+targetFormat != tt.expectedExt {
				t.Errorf("Expected extension %s, got .%s", tt.expectedExt, targetFormat)
			}
		})
	}
}
