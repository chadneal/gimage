package generate

import "fmt"

// Available models for each API
const (
	// Gemini models (AI Studio - Free tier available)
	ModelGemini25FlashImage = "gemini-2.5-flash-image"                       // Best Gemini model
	ModelGemini20FlashImage = "gemini-2.0-flash-preview-image-generation"   // Preview Gemini model
	ModelImagen3GenerateGemini = "imagen-3.0-generate-001"                   // Gemini API Imagen

	// Vertex AI models (via AI Studio API - Paid, higher quality)
	ModelImagen4Generate      = "imagen-4.0-generate-001"       // Imagen 4 Standard (AI Studio)
	ModelImagen4UltraGenerate = "imagen-4.0-ultra-generate-001" // Imagen 4 Ultra (AI Studio)
	ModelImagen4FastGenerate  = "imagen-4.0-fast-generate-001"  // Imagen 4 Fast (AI Studio)

	// Default model
	DefaultModel = ModelGemini25FlashImage
)

// ModelInfo contains metadata about a model
type ModelInfo struct {
	Name        string
	DisplayName string
	API         string // "gemini" or "vertex"
	Quality     string // "standard", "high", "premium"
	MaxSize     string // Max resolution
	Free        bool   // Free tier available
	Description string
}

// AvailableModels returns all available models
func AvailableModels() []ModelInfo {
	return []ModelInfo{
		{
			Name:        ModelGemini25FlashImage,
			DisplayName: "Gemini 2.5 Flash Image",
			API:         "gemini",
			Quality:     "high",
			MaxSize:     "1024x1024",
			Free:        true,
			Description: "Latest Gemini image model, fast and high quality (Default)",
		},
		{
			Name:        ModelGemini20FlashImage,
			DisplayName: "Gemini 2.0 Flash Image",
			API:         "gemini",
			Quality:     "high",
			MaxSize:     "1024x1024",
			Free:        true,
			Description: "Previous Gemini image model",
		},
		{
			Name:        ModelImagen3GenerateGemini,
			DisplayName: "Imagen 3 (Gemini API)",
			API:         "gemini",
			Quality:     "high",
			MaxSize:     "1536x1536",
			Free:        true,
			Description: "Imagen 3 via Gemini API",
		},
		{
			Name:        ModelImagen4Generate,
			DisplayName: "Imagen 4 Standard",
			API:         "vertex",
			Quality:     "premium",
			MaxSize:     "2048x2048",
			Free:        false,
			Description: "Imagen 4 standard quality via AI Studio (~$0.04/image)",
		},
		{
			Name:        ModelImagen4UltraGenerate,
			DisplayName: "Imagen 4 Ultra",
			API:         "vertex",
			Quality:     "premium",
			MaxSize:     "2048x2048",
			Free:        false,
			Description: "Imagen 4 ultra quality, highest fidelity (~$0.08/image)",
		},
		{
			Name:        ModelImagen4FastGenerate,
			DisplayName: "Imagen 4 Fast",
			API:         "vertex",
			Quality:     "high",
			MaxSize:     "2048x2048",
			Free:        false,
			Description: "Imagen 4 fast generation, lower latency (~$0.02/image)",
		},
	}
}

// GetModelInfo returns info for a specific model
// If a model is available on multiple APIs, returns the first match
func GetModelInfo(modelName string) (*ModelInfo, error) {
	models := AvailableModels()
	for _, m := range models {
		if m.Name == modelName {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("unknown model: %s", modelName)
}

// GetModelInfoForAPI returns info for a specific model on a specific API
func GetModelInfoForAPI(modelName, api string) (*ModelInfo, error) {
	models := AvailableModels()
	for _, m := range models {
		if m.Name == modelName && m.API == api {
			return &m, nil
		}
	}
	// If not found for the specific API, check if it exists at all
	if _, err := GetModelInfo(modelName); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("model %s is not available on %s API", modelName, api)
}

// ListModelsByAPI returns models filtered by API type
func ListModelsByAPI(api string) []ModelInfo {
	models := AvailableModels()
	filtered := []ModelInfo{}
	for _, m := range models {
		if m.API == api {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// DetectAPIFromModel determines which API to use based on model name
func DetectAPIFromModel(modelName string) (string, error) {
	if modelName == "" {
		return "gemini", nil // Default to Gemini
	}

	modelInfo, err := GetModelInfo(modelName)
	if err != nil {
		return "", err
	}

	return modelInfo.API, nil
}

// ValidateModelForAPI validates that a model is compatible with an API
func ValidateModelForAPI(modelName, api string) error {
	if modelName == "" {
		return nil // No validation needed for default
	}

	// Use API-specific lookup to check if model is available on the specified API
	_, err := GetModelInfoForAPI(modelName, api)
	return err
}
