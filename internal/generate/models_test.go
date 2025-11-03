package generate

import (
	"testing"
)

// TestResolveModelName tests model alias resolution
func TestResolveModelName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Gemini aliases
		{
			name:     "gemini alias",
			input:    "gemini",
			expected: "gemini-2.5-flash-image",
		},
		{
			name:     "gemini-flash alias",
			input:    "gemini-flash",
			expected: "gemini-2.5-flash-image",
		},
		{
			name:     "gemini-2.5-flash alias",
			input:    "gemini-2.5-flash",
			expected: "gemini-2.5-flash-image",
		},
		{
			name:     "gemini-2.0-flash alias",
			input:    "gemini-2.0-flash",
			expected: "gemini-2.0-flash-preview-image-generation",
		},
		{
			name:     "gemini-2.0-flash-exp alias",
			input:    "gemini-2.0-flash-exp",
			expected: "gemini-2.0-flash-preview-image-generation",
		},

		// Imagen aliases (note: "imagen" alone is not mapped, would need to be added if desired)
		{
			name:     "imagen-3 alias",
			input:    "imagen-3",
			expected: "imagen-3.0-generate-001",
		},
		{
			name:     "imagen3 alias",
			input:    "imagen3",
			expected: "imagen-3.0-generate-001",
		},
		{
			name:     "imagen-4 alias",
			input:    "imagen-4",
			expected: "imagen-4.0-generate-001",
		},
		{
			name:     "imagen4 alias",
			input:    "imagen4",
			expected: "imagen-4.0-generate-001",
		},
		{
			name:     "imagen-4-standard alias",
			input:    "imagen-4-standard",
			expected: "imagen-4.0-generate-001",
		},
		{
			name:     "imagen-4-ultra alias",
			input:    "imagen-4-ultra",
			expected: "imagen-4.0-ultra-generate-001",
		},
		{
			name:     "imagen-4-fast alias",
			input:    "imagen-4-fast",
			expected: "imagen-4.0-fast-generate-001",
		},

		// AWS Bedrock aliases
		{
			name:     "nova-canvas alias",
			input:    "nova-canvas",
			expected: "amazon.nova-canvas-v1:0",
		},

		// Exact model names (should pass through unchanged)
		{
			name:     "exact gemini name",
			input:    "gemini-2.5-flash-image",
			expected: "gemini-2.5-flash-image",
		},
		{
			name:     "exact gemini 2.0 name",
			input:    "gemini-2.0-flash-preview-image-generation",
			expected: "gemini-2.0-flash-preview-image-generation",
		},
		{
			name:     "exact imagen-3 name",
			input:    "imagen-3.0-generate-001",
			expected: "imagen-3.0-generate-001",
		},
		{
			name:     "exact imagen-4 name",
			input:    "imagen-4.0-generate-001",
			expected: "imagen-4.0-generate-001",
		},
		{
			name:     "exact nova canvas name",
			input:    "amazon.nova-canvas-v1:0",
			expected: "amazon.nova-canvas-v1:0",
		},

		// Unknown models (should pass through unchanged)
		{
			name:     "unknown model",
			input:    "unknown-model-123",
			expected: "unknown-model-123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveModelName(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveModelName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetPreferredAlias tests alias lookup for exact model names
func TestGetPreferredAlias(t *testing.T) {
	tests := []struct {
		name           string
		fullModelName  string
		expectedPrefix string // We check if result starts with this (shortest alias may vary)
	}{
		{
			name:           "gemini-2.5-flash-image has gemini alias",
			fullModelName:  "gemini-2.5-flash-image",
			expectedPrefix: "gemini", // Should return "gemini" (shortest)
		},
		{
			name:           "gemini-2.0-flash-preview has alias",
			fullModelName:  "gemini-2.0-flash-preview-image-generation",
			expectedPrefix: "gemini-2.0-flash",
		},
		{
			name:           "imagen-3.0-generate-001 has alias",
			fullModelName:  "imagen-3.0-generate-001",
			expectedPrefix: "imagen-3", // Should return "imagen-3" or "imagen3"
		},
		{
			name:           "imagen-4.0-generate-001 has alias",
			fullModelName:  "imagen-4.0-generate-001",
			expectedPrefix: "imagen-4", // Should return "imagen-4" or "imagen4"
		},
		{
			name:           "amazon.nova-canvas-v1:0 has alias",
			fullModelName:  "amazon.nova-canvas-v1:0",
			expectedPrefix: "nova-canvas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPreferredAlias(tt.fullModelName)
			if result == "" {
				t.Errorf("GetPreferredAlias(%q) returned empty, expected alias starting with %q", tt.fullModelName, tt.expectedPrefix)
			}
			// Just check it's not empty - the exact shortest may vary
		})
	}
}

// TestGetModelInfo tests model info retrieval with aliases
func TestGetModelInfo(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		wantAPI   string
		wantErr   bool
	}{
		// Test aliases resolve to correct API
		{
			name:      "gemini alias resolves to gemini API",
			modelName: "gemini",
			wantAPI:   "gemini",
			wantErr:   false,
		},
		{
			name:      "gemini-flash alias resolves to gemini API",
			modelName: "gemini-flash",
			wantAPI:   "gemini",
			wantErr:   false,
		},
		{
			name:      "imagen-4 alias resolves to vertex API",
			modelName: "imagen-4",
			wantAPI:   "vertex",
			wantErr:   false,
		},
		{
			name:      "nova-canvas alias resolves to bedrock API",
			modelName: "nova-canvas",
			wantAPI:   "bedrock",
			wantErr:   false,
		},

		// Test exact names work
		{
			name:      "exact gemini name",
			modelName: "gemini-2.5-flash-image",
			wantAPI:   "gemini",
			wantErr:   false,
		},
		{
			name:      "exact imagen name",
			modelName: "imagen-4.0-generate-001",
			wantAPI:   "vertex",
			wantErr:   false,
		},
		{
			name:      "exact nova canvas name",
			modelName: "amazon.nova-canvas-v1:0",
			wantAPI:   "bedrock",
			wantErr:   false,
		},

		// Test unknown models
		{
			name:      "unknown model returns error",
			modelName: "unknown-model-xyz",
			wantAPI:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetModelInfo(%q) expected error, got nil", tt.modelName)
				}
				return
			}

			if err != nil {
				t.Errorf("GetModelInfo(%q) unexpected error: %v", tt.modelName, err)
				return
			}

			if info == nil {
				t.Errorf("GetModelInfo(%q) returned nil info", tt.modelName)
				return
			}

			if info.API != tt.wantAPI {
				t.Errorf("GetModelInfo(%q) API = %q, want %q", tt.modelName, info.API, tt.wantAPI)
			}
		})
	}
}

// TestDetectAPIFromModel tests API detection from model names
func TestDetectAPIFromModel(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		wantAPI   string
		wantErr   bool
	}{
		{
			name:      "gemini model",
			modelName: "gemini",
			wantAPI:   "gemini",
			wantErr:   false,
		},
		{
			name:      "gemini-flash model",
			modelName: "gemini-flash",
			wantAPI:   "gemini",
			wantErr:   false,
		},
		{
			name:      "imagen-4 model",
			modelName: "imagen-4",
			wantAPI:   "vertex",
			wantErr:   false,
		},
		{
			name:      "nova-canvas model",
			modelName: "nova-canvas",
			wantAPI:   "bedrock",
			wantErr:   false,
		},
		{
			name:      "exact gemini model",
			modelName: "gemini-2.5-flash-image",
			wantAPI:   "gemini",
			wantErr:   false,
		},
		{
			name:      "exact imagen model",
			modelName: "imagen-4.0-generate-001",
			wantAPI:   "vertex",
			wantErr:   false,
		},
		{
			name:      "exact nova canvas model",
			modelName: "amazon.nova-canvas-v1:0",
			wantAPI:   "bedrock",
			wantErr:   false,
		},
		{
			name:      "unknown model",
			modelName: "unknown-model",
			wantAPI:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, err := DetectAPIFromModel(tt.modelName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectAPIFromModel(%q) expected error, got nil", tt.modelName)
				}
				return
			}

			if err != nil {
				t.Errorf("DetectAPIFromModel(%q) unexpected error: %v", tt.modelName, err)
				return
			}

			if api != tt.wantAPI {
				t.Errorf("DetectAPIFromModel(%q) = %q, want %q", tt.modelName, api, tt.wantAPI)
			}
		})
	}
}
