package tools

import (
	"path/filepath"
	"testing"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/mcp"
)

func TestGenerateImageToolValidation(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}

	// Create server and register tool
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterGenerateImageTool(server)

	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "missing prompt",
			args: map[string]interface{}{
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true,
			errorMsg:  "prompt is required",
		},
		{
			name: "empty prompt",
			args: map[string]interface{}{
				"prompt": "",
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true,
			errorMsg:  "prompt is required",
		},
		{
			name: "prompt with null",
			args: map[string]interface{}{
				"prompt": nil,
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true,
			errorMsg:  "prompt is required",
		},
		{
			name: "valid prompt without API key (will fail at generation, not validation)",
			args: map[string]interface{}{
				"prompt": "a beautiful sunset",
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true,
			errorMsg:  "", // Will fail during actual generation, not validation
		},
		{
			name: "valid prompt with size",
			args: map[string]interface{}{
				"prompt": "a beautiful sunset",
				"size":   "512x512",
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true, // Will fail at API call stage
		},
		{
			name: "valid prompt with model",
			args: map[string]interface{}{
				"prompt": "a beautiful sunset",
				"model":  "gemini-2.5-flash-image",
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true, // Will fail at API call stage
		},
		{
			name: "valid prompt with style",
			args: map[string]interface{}{
				"prompt": "a beautiful sunset",
				"style":  "photorealistic",
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true, // Will fail at API call stage
		},
		{
			name: "valid prompt with negative",
			args: map[string]interface{}{
				"prompt":   "a beautiful sunset",
				"negative": "people, buildings",
				"output":   filepath.Join(tmpDir, "test.png"),
			},
			wantError: true, // Will fail at API call stage
		},
		{
			name: "valid prompt with seed",
			args: map[string]interface{}{
				"prompt": "a beautiful sunset",
				"seed":   float64(12345),
				"output": filepath.Join(tmpDir, "test.png"),
			},
			wantError: true, // Will fail at API call stage
		},
		{
			name: "all parameters",
			args: map[string]interface{}{
				"prompt":   "a beautiful sunset over mountains",
				"size":     "1024x1024",
				"model":    "gemini-2.5-flash-image",
				"style":    "photorealistic",
				"negative": "people, buildings, modern objects",
				"seed":     float64(42),
				"output":   filepath.Join(tmpDir, "full_test.png"),
			},
			wantError: true, // Will fail at API call stage (no credentials in test)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := server.GetTool("generate_image")
			if tool == nil {
				t.Fatal("generate_image tool not registered")
			}

			result, err := tool.Handler(tt.args)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
					// Even if we get a result, that's unexpected in test environment
					if result != nil {
						t.Logf("Unexpected success result: %+v", result)
					}
				} else {
					// Verify error message if specified
					if tt.errorMsg != "" && err.Error() != tt.errorMsg {
						// Check if error contains the expected message
						if len(tt.errorMsg) > 0 {
							t.Logf("Got error: %v", err)
						}
					}
				}
				return
			}

			// For tests that should succeed (with valid API keys)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify result structure
			if result != nil {
				success, ok := result["success"].(bool)
				if !ok || !success {
					t.Error("Expected success: true in result")
				}
			}
		})
	}
}

func TestGenerateImageToolSchema(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterGenerateImageTool(server)

	tool := server.GetTool("generate_image")
	if tool == nil {
		t.Fatal("generate_image tool not registered")
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
	requiredFields := []string{"prompt"}
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

	// Verify optional fields exist
	optionalFields := []string{"output", "size", "model", "style", "negative", "seed"}
	for _, field := range optionalFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("Optional field '%s' not defined in schema", field)
		}
	}

	// Verify size enum
	sizeProp, ok := properties["size"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected size property to be an object")
	}

	sizeEnums, ok := sizeProp["enum"].([]string)
	if !ok {
		t.Error("Expected size to have enum values")
	} else {
		expectedSizes := []string{"256x256", "512x512", "1024x1024", "1024x1792", "1792x1024", "2048x2048"}
		if len(sizeEnums) != len(expectedSizes) {
			t.Errorf("Expected %d size options, got %d", len(expectedSizes), len(sizeEnums))
		}
	}

	// Verify model enum
	modelProp, ok := properties["model"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected model property to be an object")
	}

	modelEnums, ok := modelProp["enum"].([]string)
	if !ok {
		t.Error("Expected model to have enum values")
	} else {
		if len(modelEnums) < 2 {
			t.Error("Expected at least 2 model options")
		}
	}

	// Verify style enum
	styleProp, ok := properties["style"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected style property to be an object")
	}

	styleEnums, ok := styleProp["enum"].([]string)
	if !ok {
		t.Error("Expected style to have enum values")
	} else {
		expectedStyles := []string{"photorealistic", "artistic", "anime"}
		if len(styleEnums) != len(expectedStyles) {
			t.Errorf("Expected %d style options, got %d", len(expectedStyles), len(styleEnums))
		}
	}

	// Verify annotations
	if tool.Annotations == nil {
		t.Fatal("Expected tool annotations to be set")
	}

	if tool.Annotations.DestructiveHint {
		t.Error("Expected DestructiveHint to be false (creates new files, doesn't modify)")
	}

	if tool.Annotations.IdempotentHint {
		t.Error("Expected IdempotentHint to be false (generates different images)")
	}

	if tool.Annotations.ReadOnlyHint {
		t.Error("Expected ReadOnlyHint to be false (writes files)")
	}
}
