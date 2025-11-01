package tools

import (
	"github.com/chadneal/gimage/internal/mcp"
)

// RegisterListModelsTool registers the list_models tool
func RegisterListModelsTool(server *mcp.MCPServer) {
	tool := mcp.Tool{
		Name:        "list_models",
		Description: "List all available AI image generation models with details about their capabilities, providers, maximum resolutions, and authentication requirements. Use this to discover which models are available before generating images.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			models := []map[string]interface{}{
				{
					"name":              "gemini-2.5-flash-image",
					"provider":          "Google Gemini API",
					"description":       "Latest Gemini 2.5 Flash model optimized for fast image generation with excellent quality. Default and recommended for most use cases.",
					"max_resolution":    "1024x1792 or 1792x1024",
					"requires_api_key":  true,
					"api_key_env":       "GEMINI_API_KEY",
					"supports_styles":   true,
					"supports_negative": true,
					"supports_seed":     true,
					"free_tier":         true,
					"free_tier_limit":   "1500 requests/day",
				},
				{
					"name":              "gemini-2.0-flash-preview-image-generation",
					"provider":          "Google Gemini API",
					"description":       "Gemini 2.0 Flash preview model for image generation. Experimental features and improvements.",
					"max_resolution":    "1024x1792 or 1792x1024",
					"requires_api_key":  true,
					"api_key_env":       "GEMINI_API_KEY",
					"supports_styles":   true,
					"supports_negative": true,
					"supports_seed":     true,
					"free_tier":         true,
					"free_tier_limit":   "1500 requests/day",
				},
				{
					"name":              "imagen-3.0-generate-002",
					"provider":          "Google Vertex AI",
					"description":       "Imagen 3 model offering high-quality photorealistic image generation. Requires Vertex AI setup.",
					"max_resolution":    "1024x1024",
					"requires_api_key":  false,
					"requires_vertex":   true,
					"auth_methods":      []string{"Service Account", "Application Default Credentials", "API Key (Express Mode)"},
					"supports_styles":   true,
					"supports_negative": true,
					"supports_seed":     true,
					"free_tier":         false,
				},
				{
					"name":              "imagen-4",
					"provider":          "Google Vertex AI",
					"description":       "Latest Imagen 4 model with highest quality and most advanced features. Supports up to 2K resolution. Premium quality for professional use.",
					"max_resolution":    "2048x2048",
					"requires_api_key":  false,
					"requires_vertex":   true,
					"auth_methods":      []string{"Service Account", "Application Default Credentials", "API Key (Express Mode)"},
					"supports_styles":   true,
					"supports_negative": true,
					"supports_seed":     true,
					"free_tier":         false,
					"variants":          []string{"imagen-4-standard", "imagen-4-ultra", "imagen-4-fast"},
				},
			}

			return map[string]interface{}{
				"models": models,
				"total":  len(models),
			}, nil
		},
	}

	server.RegisterTool(tool)
}
