package generate

import (
	"fmt"
	"sort"

	"github.com/apresai/gimage/internal/config"
)

// Available models for each API
const (
	// Gemini models (AI Studio - Free tier available)
	ModelGemini25FlashImage    = "gemini-2.5-flash-image"                     // Best Gemini model
	ModelGemini20FlashImage    = "gemini-2.0-flash-preview-image-generation" // Preview Gemini model
	ModelImagen3GenerateGemini = "imagen-3.0-generate-001"                   // Gemini API Imagen

	// Vertex AI models (via AI Studio API - Paid, higher quality)
	ModelImagen4Generate      = "imagen-4.0-generate-001"       // Imagen 4 Standard (AI Studio)
	ModelImagen4UltraGenerate = "imagen-4.0-ultra-generate-001" // Imagen 4 Ultra (AI Studio)
	ModelImagen4FastGenerate  = "imagen-4.0-fast-generate-001"  // Imagen 4 Fast (AI Studio)

	// AWS Bedrock models (Paid, high quality)
	ModelNovaCanvas = "amazon.nova-canvas-v1:0" // AWS Nova Canvas

	// Default model
	DefaultModel = ModelGemini25FlashImage
)

// ModelAliases maps common model name variations to official names
var ModelAliases = map[string]string{
	"gemini-2.0-flash-exp":              "gemini-2.0-flash-preview-image-generation",
	"gemini-flash":                      "gemini-2.5-flash-image",
	"gemini":                            "gemini-2.5-flash-image",
	"imagen-3":                          "imagen-3.0-generate-001",
	"imagen3":                           "imagen-3.0-generate-001",
	"imagen-4":                          "imagen-4.0-generate-001",
	"imagen4":                           "imagen-4.0-generate-001",
	"imagen-4-standard":                 "imagen-4.0-generate-001",
	"imagen-4-ultra":                    "imagen-4.0-ultra-generate-001",
	"imagen-4-fast":                     "imagen-4.0-fast-generate-001",
	"gemini-2.5-flash":                  "gemini-2.5-flash-image",
	"gemini-2.0-flash":                  "gemini-2.0-flash-preview-image-generation",
	"nova-canvas":                       "amazon.nova-canvas-v1:0",
}

// RateLimits defines usage quotas
type RateLimits struct {
	RequestsPerMinute *int `json:"requests_per_minute,omitempty"`
	RequestsPerDay    *int `json:"requests_per_day,omitempty"`
	TokensPerMinute   *int `json:"tokens_per_minute,omitempty"`
}

// PricingInfo contains detailed pricing information for a model
type PricingInfo struct {
	// Basic Pricing
	CostPerImage   *float64 `json:"cost_per_image,omitempty"`    // USD per image (nil = free tier)
	CostPerImageHD *float64 `json:"cost_per_image_hd,omitempty"` // Higher quality pricing

	// Token-Based Pricing (for Gemini models)
	TokensPerImage   *int     `json:"tokens_per_image,omitempty"`    // Tokens consumed per image
	TokensPerImageHD *int     `json:"tokens_per_image_hd,omitempty"` // Tokens for HD images
	InputTokenCost   *float64 `json:"input_token_cost,omitempty"`    // Cost per 1M input tokens (prompt)
	OutputTokenCost  *float64 `json:"output_token_cost,omitempty"`   // Cost per 1M output tokens (image)

	// Batch Mode
	BatchModeAvailable bool     `json:"batch_mode_available"`
	BatchModeCost      *float64 `json:"batch_mode_cost,omitempty"`     // Discounted batch price
	BatchModeDiscount  *float64 `json:"batch_mode_discount,omitempty"` // Percentage discount (e.g., 0.50 = 50%)

	// Free Tier
	FreeTier      bool   `json:"free_tier"`
	FreeTierLimit string `json:"free_tier_limit,omitempty"` // e.g., "500/day" or "1500/month"
	FreeTierRPM   *int   `json:"free_tier_rpm,omitempty"`   // Requests per minute
	FreeTierRPD   *int   `json:"free_tier_rpd,omitempty"`   // Requests per day

	// Resolution & Quality Tiers
	MaxResolution     string             `json:"max_resolution"`                    // e.g., "2048x2048"
	ResolutionPricing map[string]float64 `json:"resolution_pricing,omitempty"`      // {"1024x1024": 0.04, "2048x2048": 0.08}
	SupportedSizes    []string           `json:"supported_sizes,omitempty"`

	// Rate Limits
	RateLimits RateLimits `json:"rate_limits"`

	// Billing Metadata
	BillingUnit string `json:"billing_unit"` // "per_image", "per_token", "subscription"
	Currency    string `json:"currency"`     // "USD"
	PricingTier string `json:"pricing_tier"` // "free", "standard", "premium"
	LastUpdated string `json:"last_updated"` // ISO 8601 date
}

// ModelCapabilities defines what a model can do
type ModelCapabilities struct {
	SupportsStyles         bool     `json:"supports_styles"`
	SupportsNegativePrompt bool     `json:"supports_negative_prompt"`
	SupportsSeed           bool     `json:"supports_seed"`
	SupportedStyles        []string `json:"supported_styles,omitempty"` // ["photorealistic", "artistic", "anime"]
	MaxPromptLength        int      `json:"max_prompt_length"`          // Characters or tokens
}

// ModelInfo contains metadata about a model
type ModelInfo struct {
	// Basic fields
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	API         string `json:"api"`         // "gemini" or "vertex"
	Quality     string `json:"quality"`     // "standard", "high", "premium"
	Description string `json:"description"`

	// Enhanced fields
	Pricing      PricingInfo       `json:"pricing"`
	Priority     int               `json:"priority"`      // Lower = higher priority (1 = first choice)
	RequiresAuth []string          `json:"requires_auth"` // e.g., ["GEMINI_API_KEY"]
	Capabilities ModelCapabilities `json:"capabilities"`

	// Legacy fields (kept for backward compatibility)
	MaxSize string `json:"max_size"` // Deprecated: use Pricing.MaxResolution
	Free    bool   `json:"free"`     // Deprecated: use Pricing.FreeTier
}

// Helper functions for pointer creation
func float64Ptr(f float64) *float64 { return &f }
func intPtr(i int) *int             { return &i }

// AvailableModels returns all available models with comprehensive metadata
func AvailableModels() []ModelInfo {
	return []ModelInfo{
		{
			Name:        ModelGemini25FlashImage,
			DisplayName: "Gemini 2.5 Flash Image",
			API:         "gemini",
			Quality:     "high",
			Description: "Latest Gemini image model, fast and high quality (Default)",
			Priority:    1, // Highest priority (free + good quality)
			RequiresAuth: []string{"GEMINI_API_KEY"},
			MaxSize:      "1792x1024", // Legacy field
			Free:         true,        // Legacy field

			Pricing: PricingInfo{
				// Gemini uses hybrid: per-image OR token-based
				CostPerImage:       float64Ptr(0.039),   // Standard rate after free tier
				TokensPerImage:     intPtr(1290),        // ~1290 tokens per standard image
				InputTokenCost:     float64Ptr(1.25),    // $1.25 per 1M input tokens
				OutputTokenCost:    float64Ptr(10.0),    // $10 per 1M output tokens
				BatchModeAvailable: true,
				BatchModeDiscount:  float64Ptr(0.50),    // 50% discount
				FreeTier:           true,
				FreeTierLimit:      "500 requests/day",
				FreeTierRPD:        intPtr(500),
				FreeTierRPM:        intPtr(10),
				MaxResolution:      "1792x1024",
				SupportedSizes:     []string{"256x256", "512x512", "1024x1024", "1024x1792", "1792x1024"},
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(10),
					RequestsPerDay:    intPtr(500),
					TokensPerMinute:   intPtr(250000),
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "free",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         true,
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{"photorealistic", "artistic", "anime"},
				MaxPromptLength:        480, // tokens
			},
		},

		{
			Name:        ModelGemini20FlashImage,
			DisplayName: "Gemini 2.0 Flash Image",
			API:         "gemini",
			Quality:     "high",
			Description: "Previous Gemini image model",
			Priority:    2, // Second choice (older model)
			RequiresAuth: []string{"GEMINI_API_KEY"},
			MaxSize:      "1792x1024",
			Free:         true,

			Pricing: PricingInfo{
				CostPerImage:       float64Ptr(0.039),
				TokensPerImage:     intPtr(1290),
				InputTokenCost:     float64Ptr(1.25),
				OutputTokenCost:    float64Ptr(10.0),
				BatchModeAvailable: true,
				BatchModeDiscount:  float64Ptr(0.50),
				FreeTier:           true,
				FreeTierLimit:      "500 requests/day",
				FreeTierRPD:        intPtr(500),
				FreeTierRPM:        intPtr(10),
				MaxResolution:      "1792x1024",
				SupportedSizes:     []string{"256x256", "512x512", "1024x1024", "1024x1792", "1792x1024"},
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(10),
					RequestsPerDay:    intPtr(500),
					TokensPerMinute:   intPtr(250000),
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "free",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         true,
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{"photorealistic", "artistic", "anime"},
				MaxPromptLength:        480,
			},
		},

		{
			Name:        ModelImagen3GenerateGemini,
			DisplayName: "Imagen 3 (Gemini API)",
			API:         "gemini",
			Quality:     "high",
			Description: "Imagen 3 via Gemini API",
			Priority:    6, // Lower priority (legacy)
			RequiresAuth: []string{"GEMINI_API_KEY"},
			MaxSize:      "1536x1536",
			Free:         true,

			Pricing: PricingInfo{
				CostPerImage:       float64Ptr(0.039),
				FreeTier:           true,
				FreeTierLimit:      "500 requests/day",
				FreeTierRPD:        intPtr(500),
				FreeTierRPM:        intPtr(10),
				MaxResolution:      "1536x1536",
				SupportedSizes:     []string{"256x256", "512x512", "1024x1024", "1536x1536"},
				BatchModeAvailable: true,
				BatchModeDiscount:  float64Ptr(0.50),
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(10),
					RequestsPerDay:    intPtr(500),
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "free",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         true,
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{"photorealistic", "artistic"},
				MaxPromptLength:        480,
			},
		},

		{
			Name:        ModelImagen4FastGenerate,
			DisplayName: "Imagen 4 Fast",
			API:         "vertex",
			Quality:     "high",
			Description: "Imagen 4 fast generation, lower latency",
			Priority:    3, // Third choice (cheap + fast paid option)
			RequiresAuth: []string{"VERTEX_PROJECT", "VERTEX_API_KEY or GOOGLE_APPLICATION_CREDENTIALS"},
			MaxSize:      "2048x2048",
			Free:         false,

			Pricing: PricingInfo{
				CostPerImage:       float64Ptr(0.02), // $0.02 per image (cheapest paid)
				BatchModeAvailable: true,
				BatchModeCost:      float64Ptr(0.01), // Estimated 50% discount
				FreeTier:           false,
				MaxResolution:      "2048x2048",
				SupportedSizes:     []string{"256x256", "512x512", "1024x1024", "1536x1536", "2048x2048"},
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(100), // Higher for paid tier
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "standard",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         true,
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{"photorealistic", "artistic"},
				MaxPromptLength:        480,
			},
		},

		{
			Name:        ModelImagen4Generate,
			DisplayName: "Imagen 4 Standard",
			API:         "vertex",
			Quality:     "premium",
			Description: "Imagen 4 standard quality via Vertex AI",
			Priority:    4, // Fourth choice (premium quality)
			RequiresAuth: []string{"VERTEX_PROJECT", "VERTEX_API_KEY or GOOGLE_APPLICATION_CREDENTIALS"},
			MaxSize:      "2048x2048",
			Free:         false,

			Pricing: PricingInfo{
				CostPerImage:       float64Ptr(0.04), // $0.04 per image
				BatchModeAvailable: true,
				BatchModeCost:      float64Ptr(0.02), // Estimated 50% discount
				FreeTier:           false,
				MaxResolution:      "2048x2048",
				SupportedSizes:     []string{"256x256", "512x512", "1024x1024", "1536x1536", "2048x2048"},
				ResolutionPricing: map[string]float64{
					"1024x1024": 0.04,
					"1536x1536": 0.04,
					"2048x2048": 0.04, // Same price for all resolutions
				},
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(100),
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "premium",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         true,
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{"photorealistic", "artistic"},
				MaxPromptLength:        480,
			},
		},

		{
			Name:        ModelImagen4UltraGenerate,
			DisplayName: "Imagen 4 Ultra",
			API:         "vertex",
			Quality:     "premium",
			Description: "Imagen 4 ultra quality, highest fidelity",
			Priority:    5, // Fifth choice (most expensive)
			RequiresAuth: []string{"VERTEX_PROJECT", "VERTEX_API_KEY or GOOGLE_APPLICATION_CREDENTIALS"},
			MaxSize:      "2048x2048",
			Free:         false,

			Pricing: PricingInfo{
				CostPerImage:       float64Ptr(0.06), // $0.06 per image (50% more than standard)
				BatchModeAvailable: true,
				BatchModeCost:      float64Ptr(0.03), // Estimated 50% discount
				FreeTier:           false,
				MaxResolution:      "2048x2048",
				SupportedSizes:     []string{"1024x1024", "1536x1536", "2048x2048"},
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(100),
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "premium",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         true,
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{"photorealistic", "artistic"},
				MaxPromptLength:        480,
			},
		},

		{
			Name:        ModelNovaCanvas,
			DisplayName: "AWS Nova Canvas",
			API:         "bedrock",
			Quality:     "premium",
			Description: "AWS Bedrock Nova Canvas, high quality image generation",
			Priority:    7, // After all current models
			RequiresAuth: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY or AWS_PROFILE"},
			MaxSize:      "2048x2048",
			Free:         false,

			Pricing: PricingInfo{
				CostPerImage:       float64Ptr(0.04), // $0.04 per standard image
				CostPerImageHD:     float64Ptr(0.08), // $0.08 per premium image
				BatchModeAvailable: false,
				FreeTier:           false,
				MaxResolution:      "2048x2048",
				SupportedSizes:     []string{"512x512", "768x768", "1024x1024", "1280x1280", "1536x1536", "2048x2048"},
				ResolutionPricing: map[string]float64{
					"512x512":   0.04,
					"768x768":   0.04,
					"1024x1024": 0.04,
					"1280x1280": 0.04,
					"1536x1536": 0.04,
					"2048x2048": 0.04, // Same price for all resolutions (standard quality)
				},
				RateLimits: RateLimits{
					RequestsPerMinute: intPtr(10), // 10 requests per second
				},
				BillingUnit: "per_image",
				Currency:    "USD",
				PricingTier: "standard",
				LastUpdated: "2025-11-02",
			},

			Capabilities: ModelCapabilities{
				SupportsStyles:         false, // Nova Canvas doesn't have style presets
				SupportsNegativePrompt: true,
				SupportsSeed:           true,
				SupportedStyles:        []string{}, // No predefined styles
				MaxPromptLength:        512,        // Max prompt length in characters
			},
		},
	}
}

// ResolveModelName resolves model aliases to official names
func ResolveModelName(name string) string {
	if official, ok := ModelAliases[name]; ok {
		return official
	}
	return name
}

// GetModelInfo returns info for a specific model
// Tries exact match first, then checks aliases
func GetModelInfo(modelName string) (*ModelInfo, error) {
	// Try exact match first
	models := AvailableModels()
	for i := range models {
		if models[i].Name == modelName {
			return &models[i], nil
		}
	}

	// Try alias resolution
	resolvedName := ResolveModelName(modelName)
	if resolvedName != modelName {
		for i := range models {
			if models[i].Name == resolvedName {
				return &models[i], nil
			}
		}
	}

	return nil, fmt.Errorf("unknown model: %s", modelName)
}

// GetModelInfoOrFallback tries to get model info, falls back to best available
func GetModelInfoOrFallback(modelName string) (*ModelInfo, bool, error) {
	// Try to get the requested model
	info, err := GetModelInfo(modelName)
	if err == nil {
		return info, false, nil // Found exact match, no fallback
	}

	// Model not found - select best available
	best, err := SelectBestAvailableModel("")
	if err != nil {
		return nil, false, fmt.Errorf("model %s not found and no fallback available: %w", modelName, err)
	}

	return best, true, nil // Fell back to best available
}

// GetModelInfoForAPI returns info for a specific model on a specific API
func GetModelInfoForAPI(modelName, api string) (*ModelInfo, error) {
	// First try to resolve via GetModelInfo (handles aliases)
	modelInfo, err := GetModelInfo(modelName)
	if err != nil {
		return nil, err
	}

	// Check if it's on the requested API
	if modelInfo.API != api {
		return nil, fmt.Errorf("model %s is not available on %s API (found on %s)", modelName, api, modelInfo.API)
	}

	return modelInfo, nil
}

// GetPreferredAlias returns the shortest/most user-friendly alias for a model
// Returns empty string if no alias exists
func GetPreferredAlias(fullModelName string) string {
	// Find all aliases that map to this model
	var aliases []string
	for alias, target := range ModelAliases {
		if target == fullModelName {
			aliases = append(aliases, alias)
		}
	}

	if len(aliases) == 0 {
		return "" // No alias found
	}

	// Return the shortest alias (most user-friendly)
	shortest := aliases[0]
	for _, alias := range aliases {
		if len(alias) < len(shortest) {
			shortest = alias
		}
	}
	return shortest
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

// SelectBestAvailableModel returns the highest priority model that has valid credentials
func SelectBestAvailableModel(preferredAPI string) (*ModelInfo, error) {
	hasGemini := config.HasGeminiCredentials()
	hasVertex := config.HasVertexCredentials()
	hasBedrock := config.HasBedrockCredentials()

	if !hasGemini && !hasVertex && !hasBedrock {
		return nil, fmt.Errorf("no API credentials configured. Run 'gimage auth gemini', 'gimage auth vertex', or 'gimage auth bedrock'")
	}

	// Get all models and sort by priority
	models := AvailableModels()
	sort.Slice(models, func(i, j int) bool {
		return models[i].Priority < models[j].Priority
	})

	// Find the first model that matches credentials and optional API preference
	for i := range models {
		model := &models[i]

		// Check API preference if specified
		if preferredAPI != "" && model.API != preferredAPI {
			continue
		}

		// Check if we have credentials for this model's API
		if model.API == "gemini" && hasGemini {
			return model, nil
		}
		if model.API == "vertex" && hasVertex {
			return model, nil
		}
		if model.API == "bedrock" && hasBedrock {
			return model, nil
		}
	}

	return nil, fmt.Errorf("no available models for configured credentials")
}

// GetEstimatedCost calculates estimated cost for a generation request
func GetEstimatedCost(modelInfo *ModelInfo, size string, quantity int) (cost float64, tokens int, explanation string) {
	pricing := modelInfo.Pricing

	// Check if free tier applies
	if pricing.FreeTier && pricing.FreeTierRPD != nil && quantity <= *pricing.FreeTierRPD {
		// For token-based models, still calculate token usage
		if pricing.TokensPerImage != nil {
			tokens = *pricing.TokensPerImage * quantity
			return 0.0, tokens, fmt.Sprintf("FREE tier (%d/%d daily limit, ~%d tokens)", quantity, *pricing.FreeTierRPD, tokens)
		}
		return 0.0, 0, fmt.Sprintf("FREE tier (%d/%d daily limit)", quantity, *pricing.FreeTierRPD)
	}

	// Calculate token usage for Gemini models
	if pricing.TokensPerImage != nil {
		tokens = *pricing.TokensPerImage * quantity
	}

	// Check for resolution-specific pricing
	if pricing.ResolutionPricing != nil {
		if resPrice, ok := pricing.ResolutionPricing[size]; ok {
			cost = resPrice * float64(quantity)
			if tokens > 0 {
				return cost, tokens, fmt.Sprintf("$%.4f per %s image × %d = $%.4f (~%d tokens)", resPrice, size, quantity, cost, tokens)
			}
			return cost, tokens, fmt.Sprintf("$%.4f per %s image × %d = $%.4f", resPrice, size, quantity, cost)
		}
	}

	// Use base cost per image
	if pricing.CostPerImage != nil {
		cost = *pricing.CostPerImage * float64(quantity)
		if tokens > 0 {
			return cost, tokens, fmt.Sprintf("$%.4f per image × %d = $%.4f (~%d tokens)", *pricing.CostPerImage, quantity, cost, tokens)
		}
		return cost, tokens, fmt.Sprintf("$%.4f per image × %d = $%.4f", *pricing.CostPerImage, quantity, cost)
	}

	return 0.0, tokens, "Pricing information not available"
}

// FormatPricingDisplay returns a human-readable pricing summary
func FormatPricingDisplay(modelInfo *ModelInfo) string {
	p := modelInfo.Pricing

	if p.FreeTier && p.FreeTierLimit != "" {
		if p.CostPerImage != nil {
			return fmt.Sprintf("FREE (%s), then $%.4f/image", p.FreeTierLimit, *p.CostPerImage)
		}
		return fmt.Sprintf("FREE (%s)", p.FreeTierLimit)
	}

	if p.CostPerImage != nil {
		if p.BatchModeCost != nil {
			return fmt.Sprintf("$%.4f/image (batch: $%.4f)", *p.CostPerImage, *p.BatchModeCost)
		}
		return fmt.Sprintf("$%.4f/image", *p.CostPerImage)
	}

	return "Pricing varies"
}
