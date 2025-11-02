package tools

import (
	"sort"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/generate"
	"github.com/apresai/gimage/internal/mcp"
)

// RegisterListModelsTool registers the list_models tool
func RegisterListModelsTool(server *mcp.MCPServer) {
	tool := mcp.Tool{
		Name:        "list_models",
		Description: "List all available AI image generation models with detailed pricing, capabilities, and authentication requirements. Shows which models are currently accessible based on configured credentials. Returns comprehensive pricing information including cost per image, token usage, free tiers, and batch pricing.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			// Get all models from generate package (single source of truth)
			allModels := generate.AvailableModels()

			// Check credentials
			hasGemini := config.HasGeminiCredentials()
			hasVertex := config.HasVertexCredentials()

			// Build model list with availability
			models := []map[string]interface{}{}
			for _, m := range allModels {
				available := false
				if m.API == "gemini" && hasGemini {
					available = true
				} else if m.API == "vertex" && hasVertex {
					available = true
				}

				// Convert pricing info to map
				pricingMap := map[string]interface{}{
					"billing_unit": m.Pricing.BillingUnit,
					"currency":     m.Pricing.Currency,
					"pricing_tier": m.Pricing.PricingTier,
					"free_tier":    m.Pricing.FreeTier,
				}

				if m.Pricing.CostPerImage != nil {
					pricingMap["cost_per_image"] = *m.Pricing.CostPerImage
				}
				if m.Pricing.TokensPerImage != nil {
					pricingMap["tokens_per_image"] = *m.Pricing.TokensPerImage
				}
				if m.Pricing.FreeTierLimit != "" {
					pricingMap["free_tier_limit"] = m.Pricing.FreeTierLimit
				}
				if m.Pricing.BatchModeAvailable {
					pricingMap["batch_mode_available"] = true
					if m.Pricing.BatchModeCost != nil {
						pricingMap["batch_cost_per_image"] = *m.Pricing.BatchModeCost
					}
				}

				modelData := map[string]interface{}{
					"name":            m.Name,
					"display_name":    m.DisplayName,
					"api":             m.API,
					"quality":         m.Quality,
					"description":     m.Description,
					"priority":        m.Priority,
					"available":       available,
					"requires_auth":   m.RequiresAuth,
					"max_resolution":  m.Pricing.MaxResolution,
					"supported_sizes": m.Pricing.SupportedSizes,

					// Pricing information
					"pricing":         pricingMap,
					"pricing_summary": generate.FormatPricingDisplay(&m),

					// Capabilities
					"supports_styles":         m.Capabilities.SupportsStyles,
					"supports_negative_prompt": m.Capabilities.SupportsNegativePrompt,
					"supports_seed":           m.Capabilities.SupportsSeed,
					"supported_styles":        m.Capabilities.SupportedStyles,
					"max_prompt_length":       m.Capabilities.MaxPromptLength,
				}

				models = append(models, modelData)
			}

			// Sort by priority
			sort.Slice(models, func(i, j int) bool {
				return models[i]["priority"].(int) < models[j]["priority"].(int)
			})

			// Get default model
			var defaultModelName string
			var defaultModelDisplay string
			var defaultModelPricing string
			if defaultModel, err := generate.SelectBestAvailableModel(""); err == nil {
				defaultModelName = defaultModel.Name
				defaultModelDisplay = defaultModel.DisplayName
				defaultModelPricing = generate.FormatPricingDisplay(defaultModel)
			}

			return map[string]interface{}{
				"models": models,
				"total":  len(models),
				"credentials": map[string]interface{}{
					"gemini_configured": hasGemini,
					"vertex_configured": hasVertex,
				},
				"default_model": map[string]interface{}{
					"name":           defaultModelName,
					"display_name":   defaultModelDisplay,
					"pricing_summary": defaultModelPricing,
				},
				"pricing_note": "Costs shown are in USD. Free tier limits reset daily. Token-based models (Gemini) charge based on image complexity (~1290 tokens/image for standard photos). Batch mode offers ~50% discount for async processing.",
				"recommendations": map[string]interface{}{
					"free_users": "gemini-2.5-flash-image (500 FREE images/day, excellent quality)",
					"paid_users": "imagen-4.0-fast-generate-001 ($0.02/image, fastest paid option)",
					"max_quality": "imagen-4.0-ultra-generate-001 ($0.06/image, highest fidelity)",
				},
			}, nil
		},
	}

	server.RegisterTool(tool)
}
