package tui

import (
	"testing"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/generate"
)

// TestModelDetection verifies that model detection works correctly in TUI context
func TestModelDetection(t *testing.T) {
	tests := []struct {
		modelName   string
		expectedAPI string
	}{
		{"gemini-2.5-flash-image", "gemini"},
		{"gemini-2.0-flash-preview-image-generation", "gemini"},
		{"imagen-4.0-generate-001", "vertex"},
		{"amazon.nova-canvas-v1:0", "bedrock"},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			api, err := generate.DetectAPIFromModel(tt.modelName)
			if err != nil {
				t.Errorf("DetectAPIFromModel(%s) returned error: %v", tt.modelName, err)
				return
			}
			if api != tt.expectedAPI {
				t.Errorf("DetectAPIFromModel(%s) = %s, want %s", tt.modelName, api, tt.expectedAPI)
			}

			// Also test GetModelInfo
			modelInfo, err := generate.GetModelInfo(tt.modelName)
			if err != nil {
				t.Errorf("GetModelInfo(%s) returned error: %v", tt.modelName, err)
				return
			}
			if modelInfo.API != tt.expectedAPI {
				t.Errorf("GetModelInfo(%s).API = %s, want %s", tt.modelName, modelInfo.API, tt.expectedAPI)
			}
		})
	}
}

// TestTUIGenerationPath tests the exact path the TUI takes
func TestTUIGenerationPath(t *testing.T) {
	// Load config like TUI does
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Logf("Warning: config load failed: %v", err)
		cfg = &config.Config{}
	}

	// Test model that TUI uses
	modelName := "gemini-2.5-flash-image"

	// Step 1: Test what TUI does - DetectAPIFromModel
	api, err := generate.DetectAPIFromModel(modelName)
	if err != nil {
		t.Fatalf("DetectAPIFromModel failed: %v", err)
	}

	t.Logf("Model: %s -> API: %s", modelName, api)

	// Step 2: Try to create client (without actually calling it)
	switch api {
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			t.Logf("Gemini API key not configured - would fail in real usage")
		} else {
			t.Logf("Gemini API key IS configured - should work")
			// Try creating client
			client, err := generate.NewGeminiClient(cfg.GeminiAPIKey)
			if err != nil {
				t.Errorf("Failed to create Gemini client: %v", err)
			} else {
				client.Close()
				t.Logf("Successfully created Gemini client")
			}
		}
	case "vertex":
		t.Logf("Would try Vertex AI - this might be the problem!")
	case "bedrock":
		t.Logf("Would try AWS Bedrock")
	default:
		t.Errorf("Unknown API: %s", api)
	}

	// Step 4: Log exact model string bytes to check for hidden characters
	t.Logf("Model name bytes: %x", []byte(modelName))
	t.Logf("Model name length: %d", len(modelName))

	// Step 5: Check if there's any alias resolution happening
	resolved := generate.ResolveModelName(modelName)
	if resolved != modelName {
		t.Logf("WARNING: Model name was resolved from %s to %s", modelName, resolved)
	}
}

// TestModelListConsistency checks if model list is consistent
func TestModelListConsistency(t *testing.T) {
	models := generate.AvailableModels()

	for _, m := range models {
		// Check if each model's API detection matches its declared API
		detectedAPI, err := generate.DetectAPIFromModel(m.Name)
		if err != nil {
			t.Errorf("Model %s: DetectAPIFromModel failed: %v", m.Name, err)
			continue
		}

		if detectedAPI != m.API {
			t.Errorf("Model %s: API mismatch - declared: %s, detected: %s",
				m.Name, m.API, detectedAPI)
		}

		t.Logf("✓ Model %s: API=%s (consistent)", m.Name, m.API)
	}
}

// TestExactTUIModelSelection simulates exact TUI model selection
func TestExactTUIModelSelection(t *testing.T) {
	// Create a GenerateFlowModel like TUI does
	m := NewGenerateFlowModel()

	// Check what models are loaded
	t.Logf("TUI loaded %d models:", len(m.models))
	for i, model := range m.models {
		api, _ := generate.DetectAPIFromModel(model.name)
		t.Logf("  [%d] name=%q displayName=%q -> API=%s",
			i, model.name, model.displayName, api)
	}

	// Check if first model (default selection) is correct
	if len(m.models) > 0 {
		firstModel := m.models[0]
		t.Logf("\nDefault selected model: name=%q", firstModel.name)

		// This is what gets passed to generation
		api, err := generate.DetectAPIFromModel(firstModel.name)
		if err != nil {
			t.Errorf("Default model detection failed: %v", err)
		} else {
			t.Logf("Default model would use API: %s", api)
		}
	}
}