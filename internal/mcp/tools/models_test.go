package tools

import (
	"testing"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/mcp"
)

func TestListModelsTool(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	// Test with no arguments (list_models doesn't need any)
	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify result structure
	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected models array in result")
	}

	// Should have at least 4 models (2 Gemini + 2 Vertex)
	if len(models) < 4 {
		t.Errorf("Expected at least 4 models, got %d", len(models))
	}

	// Verify total count
	total, ok := result["total"].(int)
	if !ok {
		t.Error("Expected total in result")
	} else if total != len(models) {
		t.Errorf("Total (%d) doesn't match models length (%d)", total, len(models))
	}

	// Verify each model has required fields
	requiredFields := []string{
		"name",
		"provider",
		"description",
		"max_resolution",
		"requires_api_key",
		"supports_styles",
		"supports_negative",
		"supports_seed",
	}

	for i, model := range models {
		for _, field := range requiredFields {
			if _, exists := model[field]; !exists {
				t.Errorf("Model %d missing required field: %s", i, field)
			}
		}

		// Verify specific field types
		if name, ok := model["name"].(string); !ok || name == "" {
			t.Errorf("Model %d has invalid name", i)
		}

		if provider, ok := model["provider"].(string); !ok || provider == "" {
			t.Errorf("Model %d has invalid provider", i)
		}

		if _, ok := model["requires_api_key"].(bool); !ok {
			t.Errorf("Model %d has invalid requires_api_key", i)
		}
	}

	// Verify we have both Gemini and Vertex models
	hasGemini := false
	hasVertex := false

	for _, model := range models {
		provider := model["provider"].(string)
		if provider == "Google Gemini API" {
			hasGemini = true
		}
		if provider == "Google Vertex AI" {
			hasVertex = true
		}
	}

	if !hasGemini {
		t.Error("Expected at least one Gemini model")
	}

	if !hasVertex {
		t.Error("Expected at least one Vertex model")
	}
}

func TestListModelsToolSpecificModels(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	models, _ := result["models"].([]map[string]interface{})

	// Test for specific known models
	expectedModels := map[string]string{
		"gemini-2.5-flash-image":                     "Google Gemini API",
		"gemini-2.0-flash-preview-image-generation":  "Google Gemini API",
		"imagen-3.0-generate-002":                    "Google Vertex AI",
		"imagen-4":                                   "Google Vertex AI",
	}

	for expectedName, expectedProvider := range expectedModels {
		found := false
		for _, model := range models {
			if model["name"] == expectedName {
				found = true
				if model["provider"] != expectedProvider {
					t.Errorf("Model %s has wrong provider: expected %s, got %s",
						expectedName, expectedProvider, model["provider"])
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s not found", expectedName)
		}
	}
}

func TestListModelsToolFreeTier(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	models, _ := result["models"].([]map[string]interface{})

	// Verify Gemini models have free tier info
	for _, model := range models {
		if model["provider"] == "Google Gemini API" {
			freeTier, ok := model["free_tier"].(bool)
			if !ok {
				t.Errorf("Model %s missing free_tier field", model["name"])
				continue
			}

			if freeTier {
				// Should have free_tier_limit
				if _, exists := model["free_tier_limit"]; !exists {
					t.Errorf("Model %s has free_tier=true but no free_tier_limit", model["name"])
				}
			}
		}
	}
}

func TestListModelsToolResolutionInfo(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	models, _ := result["models"].([]map[string]interface{})

	// Verify specific resolution capabilities
	for _, model := range models {
		maxRes, ok := model["max_resolution"].(string)
		if !ok || maxRes == "" {
			t.Errorf("Model %s has invalid max_resolution", model["name"])
			continue
		}

		// Imagen 4 should support 2K
		if model["name"] == "imagen-4" {
			if maxRes != "2048x2048" {
				t.Errorf("Imagen 4 should support 2048x2048, got: %s", maxRes)
			}
		}

		// Gemini models should support at least 1792
		if model["provider"] == "Google Gemini API" {
			if maxRes != "1792x1024 or 1024x1792" && maxRes != "1024x1792 or 1792x1024" {
				t.Errorf("Gemini model has unexpected max resolution: %s", maxRes)
			}
		}
	}
}

func TestListModelsToolCapabilities(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	models, _ := result["models"].([]map[string]interface{})

	// All models should support styles, negative prompts, and seeds
	for _, model := range models {
		capabilities := []string{"supports_styles", "supports_negative", "supports_seed"}
		for _, cap := range capabilities {
			supported, ok := model[cap].(bool)
			if !ok {
				t.Errorf("Model %s missing capability: %s", model["name"], cap)
				continue
			}

			if !supported {
				t.Errorf("Model %s should support %s", model["name"], cap)
			}
		}
	}
}

func TestListModelsToolSchema(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	// Verify schema structure
	schema := tool.InputSchema
	if schema["type"] != "object" {
		t.Error("Expected type 'object' in schema")
	}

	// list_models takes no parameters
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		properties = make(map[string]interface{})
	}

	if len(properties) > 0 {
		t.Error("list_models should not have any parameters")
	}

	// Should have no required fields
	required, ok := schema["required"].([]string)
	if ok && len(required) > 0 {
		t.Error("list_models should not have required parameters")
	}
}
