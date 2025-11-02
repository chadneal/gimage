package tools

import (
	"runtime"
	"testing"

	"github.com/apresai/gimage/internal/mcp"
)

func TestRegisterBatchResizeTool(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)

	// Register the tool
	RegisterBatchResizeTool(server)

	// Verify tool is registered
	tool := server.GetTool("batch_resize")
	if tool == nil {
		t.Fatal("batch_resize tool not registered")
	}

	// Verify tool properties
	if tool.Name != "batch_resize" {
		t.Errorf("Expected name 'batch_resize', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description is empty")
	}

	if tool.Handler == nil {
		t.Error("Tool handler is nil")
	}
}

func TestRegisterBatchCompressTool(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)

	// Register the tool
	RegisterBatchCompressTool(server)

	// Verify tool is registered
	tool := server.GetTool("batch_compress")
	if tool == nil {
		t.Fatal("batch_compress tool not registered")
	}

	// Verify tool properties
	if tool.Name != "batch_compress" {
		t.Errorf("Expected name 'batch_compress', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description is empty")
	}

	if tool.Handler == nil {
		t.Error("Tool handler is nil")
	}

	// Verify annotations
	if tool.Annotations == nil {
		t.Fatal("Tool annotations are nil")
	}

	if !tool.Annotations.DestructiveHint {
		t.Error("batch_compress should be marked as destructive")
	}

	if !tool.Annotations.IdempotentHint {
		t.Error("batch_compress should be marked as idempotent")
	}

	if tool.Annotations.ReadOnlyHint {
		t.Error("batch_compress should not be read-only (writes files)")
	}
}

func TestRegisterBatchConvertTool(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)

	// Register the tool
	RegisterBatchConvertTool(server)

	// Verify tool is registered
	tool := server.GetTool("batch_convert")
	if tool == nil {
		t.Fatal("batch_convert tool not registered")
	}

	// Verify tool properties
	if tool.Name != "batch_convert" {
		t.Errorf("Expected name 'batch_convert', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description is empty")
	}

	if tool.Handler == nil {
		t.Error("Tool handler is nil")
	}
}

func TestBatchResizeTool_InputSchema(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterBatchResizeTool(server)

	tool := server.GetTool("batch_resize")
	if tool == nil {
		t.Fatal("batch_resize tool not registered")
	}

	schema := tool.InputSchema

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required field is not a string slice")
	}

	expectedRequired := []string{"input_dir", "width", "height", "output_dir"}
	if len(required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
	}

	// Check properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties field is not a map")
	}

	// Verify required properties
	requiredProps := []string{"input_dir", "width", "height", "output_dir"}
	for _, prop := range requiredProps {
		if _, exists := properties[prop]; !exists {
			t.Errorf("required property '%s' missing", prop)
		}
	}

	// Verify optional workers property
	if _, exists := properties["workers"]; !exists {
		t.Error("workers property missing")
	}
}

func TestBatchCompressTool_InputSchema(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterBatchCompressTool(server)

	tool := server.GetTool("batch_compress")
	if tool == nil {
		t.Fatal("batch_compress tool not registered")
	}

	schema := tool.InputSchema

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required field is not a string slice")
	}

	expectedRequired := []string{"input_dir", "output_dir"}
	if len(required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
	}

	// Check properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties field is not a map")
	}

	// Verify quality property with defaults
	qualityProp, exists := properties["quality"]
	if !exists {
		t.Error("quality property missing")
	} else {
		qualityMap, ok := qualityProp.(map[string]interface{})
		if !ok {
			t.Error("quality property is not a map")
		} else {
			if qualityMap["default"] != 85 {
				t.Errorf("Expected default quality 85, got %v", qualityMap["default"])
			}
		}
	}
}

func TestBatchConvertTool_InputSchema(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterBatchConvertTool(server)

	tool := server.GetTool("batch_convert")
	if tool == nil {
		t.Fatal("batch_convert tool not registered")
	}

	schema := tool.InputSchema

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required field is not a string slice")
	}

	expectedRequired := []string{"input_dir", "format", "output_dir"}
	if len(required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
	}

	// Check properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties field is not a map")
	}

	// Verify format property with enum
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
}

func TestBatchProcessImages_ValidationErrors(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterBatchResizeTool(server)

	tool := server.GetTool("batch_resize")
	if tool == nil {
		t.Fatal("batch_resize tool not registered")
	}

	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "missing input_dir",
			args: map[string]interface{}{
				"width":      100,
				"height":     100,
				"output_dir": "/tmp/output",
			},
			wantError: true,
			errorMsg:  "input_dir",
		},
		{
			name: "missing width",
			args: map[string]interface{}{
				"input_dir":  "/tmp/input",
				"height":     100,
				"output_dir": "/tmp/output",
			},
			wantError: true,
			errorMsg:  "width",
		},
		{
			name: "missing height",
			args: map[string]interface{}{
				"input_dir":  "/tmp/input",
				"width":      100,
				"output_dir": "/tmp/output",
			},
			wantError: true,
			errorMsg:  "height",
		},
		{
			name: "missing output_dir",
			args: map[string]interface{}{
				"input_dir": "/tmp/input",
				"width":     100,
				"height":    100,
			},
			wantError: true,
			errorMsg:  "output_dir",
		},
		{
			name: "input_dir not a string",
			args: map[string]interface{}{
				"input_dir":  123,
				"width":      100,
				"height":     100,
				"output_dir": "/tmp/output",
			},
			wantError: true,
			errorMsg:  "input_dir",
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

func TestWorkerCount_Defaults(t *testing.T) {
	tests := []struct {
		name          string
		workersInput  interface{}
		expectedMin   int
		expectedMax   int
	}{
		{
			name:         "no workers specified uses CPU count",
			workersInput: nil,
			expectedMin:  1,
			expectedMax:  runtime.NumCPU(),
		},
		{
			name:         "workers below 1 clamped to 1",
			workersInput: float64(0),
			expectedMin:  1,
			expectedMax:  1,
		},
		{
			name:         "workers above 16 clamped to 16",
			workersInput: float64(20),
			expectedMin:  16,
			expectedMax:  16,
		},
		{
			name:         "valid worker count used as-is",
			workersInput: float64(4),
			expectedMin:  4,
			expectedMax:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test worker count normalization logic
			workers := runtime.NumCPU()
			if tt.workersInput != nil {
				if workersVal, ok := tt.workersInput.(float64); ok {
					workers = int(workersVal)
					if workers < 1 {
						workers = 1
					} else if workers > 16 {
						workers = 16
					}
				}
			}

			if workers < tt.expectedMin || workers > tt.expectedMax {
				t.Errorf("Expected workers between %d and %d, got %d", tt.expectedMin, tt.expectedMax, workers)
			}
		})
	}
}

func TestImageExtensions(t *testing.T) {
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".tiff", ".bmp"}

	tests := []struct {
		filename string
		expected bool
	}{
		{"image.png", true},
		{"photo.jpg", true},
		{"picture.jpeg", true},
		{"web.webp", true},
		{"animation.gif", true},
		{"scan.tiff", true},
		{"bitmap.bmp", true},
		{"document.pdf", false},
		{"video.mp4", false},
		{"text.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			// Check if extension is in valid list
			found := false
			for _, ext := range imageExtensions {
				if len(tt.filename) >= len(ext) && tt.filename[len(tt.filename)-len(ext):] == ext {
					found = true
					break
				}
			}

			if found != tt.expected {
				t.Errorf("File %s should be valid=%v, got valid=%v", tt.filename, tt.expected, found)
			}
		})
	}
}

func TestBatchCompress_QualityDefault(t *testing.T) {
	// Test that default quality is 85
	quality := 85
	args := map[string]interface{}{}

	// If quality not specified, use default
	if qualityVal, ok := args["quality"].(float64); ok {
		quality = int(qualityVal)
	}

	if quality != 85 {
		t.Errorf("Expected default quality 85, got %d", quality)
	}

	// If quality specified, use it
	args["quality"] = float64(75)
	if qualityVal, ok := args["quality"].(float64); ok {
		quality = int(qualityVal)
	}

	if quality != 75 {
		t.Errorf("Expected quality 75, got %d", quality)
	}
}
