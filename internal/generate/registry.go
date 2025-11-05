// Package generate provides a unified registry for managing AI model providers,
// clients, and authentication across multiple cloud platforms.
package generate

import (
	"context"
	"fmt"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/pkg/models"
)

// ImageGenerator is the common interface for all image generation clients
type ImageGenerator interface {
	GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error)
	Close() error
}

// ClientType represents the type of client implementation to use
type ClientType string

const (
	ClientTypeREST ClientType = "rest"
	ClientTypeSDK  ClientType = "sdk"
)

// ClientFactory is a function that creates an ImageGenerator client
type ClientFactory func(cfg *config.Config) (ImageGenerator, error)

// ModelProvider represents a specific model's configuration and routing information
type ModelProvider struct {
	// Model identification
	ModelID     string   // Exact model ID used by the API
	DisplayName string   // Human-friendly name
	Aliases     []string // Alternative names users might use

	// API configuration
	API        string     // API provider (gemini, vertex, bedrock)
	ClientType ClientType // Which client implementation to use

	// API-specific model name (if different from ModelID)
	APIModelName map[string]string // Maps API to specific model name format

	// Required credentials
	RequiredCredentials []string // List of required env vars or config keys

	// Capabilities
	Capabilities ModelCapabilities

	// Pricing
	Pricing PricingInfo

	// Client factory
	ClientFactory ClientFactory
}

// Registry manages all model providers and their configurations
type Registry struct {
	providers map[string]*ModelProvider // Key is ModelID
	aliases   map[string]string         // Maps aliases to ModelID
}

// Global registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new model registry
func NewRegistry() *Registry {
	r := &Registry{
		providers: make(map[string]*ModelProvider),
		aliases:   make(map[string]string),
	}
	r.registerAllProviders()
	return r
}

// GetRegistry returns the global registry instance
func GetRegistry() *Registry {
	return globalRegistry
}

// registerAllProviders registers all known model providers
func (r *Registry) registerAllProviders() {
	// Gemini Models
	r.Register(&ModelProvider{
		ModelID:     "gemini-2.5-flash-image",
		DisplayName: "Gemini 2.5 Flash Image",
		Aliases:     []string{"gemini", "gemini-flash", "flash"},
		API:         "gemini",
		ClientType:  ClientTypeREST, // REST client works, SDK is broken
		APIModelName: map[string]string{
			"gemini": "gemini-2.5-flash-image",
		},
		RequiredCredentials: []string{"GEMINI_API_KEY"},
		Capabilities: ModelCapabilities{
			SupportsStyles:         true,
			SupportsNegativePrompt: true,
			SupportsSeed:           true,
			SupportedStyles:        []string{"photorealistic", "artistic", "anime"},
			MaxPromptLength:        480,
		},
		Pricing: PricingInfo{
			CostPerImage:  float64Ptr(0.039),
			FreeTier:      true,
			FreeTierLimit: "500 requests/day",
		},
		ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
			if cfg.GeminiAPIKey == "" {
				return nil, fmt.Errorf("GEMINI_API_KEY is required")
			}
			// Always use REST client for Gemini (SDK has bugs)
			return NewGeminiRESTClient(cfg.GeminiAPIKey)
		},
	})

	r.Register(&ModelProvider{
		ModelID:     "gemini-2.0-flash-preview-image-generation",
		DisplayName: "Gemini 2.0 Flash Preview",
		Aliases:     []string{"gemini-2", "gemini-preview"},
		API:         "gemini",
		ClientType:  ClientTypeREST,
		APIModelName: map[string]string{
			"gemini": "gemini-2.0-flash-preview-image-generation",
		},
		RequiredCredentials: []string{"GEMINI_API_KEY"},
		Capabilities: ModelCapabilities{
			SupportsStyles:         true,
			SupportsNegativePrompt: true,
			SupportsSeed:           true,
			SupportedStyles:        []string{"photorealistic", "artistic", "anime"},
			MaxPromptLength:        480,
		},
		Pricing: PricingInfo{
			CostPerImage:  float64Ptr(0.0),
			FreeTier:      true,
			FreeTierLimit: "unlimited during preview",
		},
		ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
			if cfg.GeminiAPIKey == "" {
				return nil, fmt.Errorf("GEMINI_API_KEY is required")
			}
			return NewGeminiRESTClient(cfg.GeminiAPIKey)
		},
	})

	// Vertex AI Models
	r.Register(&ModelProvider{
		ModelID:     "imagen-4.0-generate-001",
		DisplayName: "Imagen 4",
		Aliases:     []string{"imagen", "imagen-4", "imagen4"},
		API:         "vertex",
		ClientType:  ClientTypeREST, // Can use REST or SDK
		APIModelName: map[string]string{
			"vertex": "imagen-4.0-generate-001",
		},
		RequiredCredentials: []string{"VERTEX_API_KEY", "VERTEX_PROJECT", "VERTEX_LOCATION"},
		Capabilities: ModelCapabilities{
			SupportsStyles:         true,
			SupportsNegativePrompt: true,
			SupportsSeed:           true,
			SupportedStyles:        []string{"photorealistic", "artistic", "anime", "digital-art", "photo"},
			MaxPromptLength:        2000,
		},
		Pricing: PricingInfo{
			CostPerImage: float64Ptr(0.04),
			FreeTier:     false,
		},
		ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
			// Check for Express Mode (API key) first
			if cfg.VertexAPIKey != "" {
				return NewVertexRESTClient(cfg.VertexAPIKey, cfg.VertexProject, cfg.VertexLocation)
			}
			// Fall back to SDK with service account
			ctx := context.Background()
			return NewVertexSDKClient(ctx, cfg.VertexProject, cfg.VertexLocation)
		},
	})

	r.Register(&ModelProvider{
		ModelID:     "imagen-3.0-generate-001",
		DisplayName: "Imagen 3",
		Aliases:     []string{"imagen-3", "imagen3"},
		API:         "vertex",
		ClientType:  ClientTypeREST,
		APIModelName: map[string]string{
			"vertex": "imagen-3.0-generate-001",
		},
		RequiredCredentials: []string{"VERTEX_API_KEY", "VERTEX_PROJECT", "VERTEX_LOCATION"},
		Capabilities: ModelCapabilities{
			SupportsStyles:         true,
			SupportsNegativePrompt: true,
			SupportsSeed:           true,
			SupportedStyles:        []string{"photorealistic", "artistic"},
			MaxPromptLength:        2000,
		},
		Pricing: PricingInfo{
			CostPerImage: float64Ptr(0.02),
			FreeTier:     false,
		},
		ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
			if cfg.VertexAPIKey != "" {
				return NewVertexRESTClient(cfg.VertexAPIKey, cfg.VertexProject, cfg.VertexLocation)
			}
			ctx := context.Background()
			return NewVertexSDKClient(ctx, cfg.VertexProject, cfg.VertexLocation)
		},
	})

	// AWS Bedrock Models
	r.Register(&ModelProvider{
		ModelID:     "amazon.nova-canvas-v1:0",
		DisplayName: "Nova Canvas",
		Aliases:     []string{"nova", "nova-canvas", "bedrock"},
		API:         "bedrock",
		ClientType:  ClientTypeREST,
		APIModelName: map[string]string{
			"bedrock": "amazon.nova-canvas-v1:0",
		},
		RequiredCredentials: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION"},
		Capabilities: ModelCapabilities{
			SupportsStyles:         true,
			SupportsNegativePrompt: true,
			SupportsSeed:           true,
			SupportedStyles:        []string{"photorealistic", "artistic", "anime"},
			MaxPromptLength:        4096,
		},
		Pricing: PricingInfo{
			CostPerImage: float64Ptr(0.08),
			FreeTier:     false,
		},
		ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
			// Try bearer token first (REST)
			if cfg.AWSBedrockAPIKey != "" {
				return NewBedrockRESTClient(cfg.AWSBedrockAPIKey, cfg.AWSRegion)
			}
			// Fall back to SDK with IAM
			ctx := context.Background()
			return NewBedrockSDKClient(ctx, cfg.AWSRegion)
		},
	})

	r.Register(&ModelProvider{
		ModelID:     "amazon.nova-lite-v1:0",
		DisplayName: "Nova Lite",
		Aliases:     []string{"nova-lite"},
		API:         "bedrock",
		ClientType:  ClientTypeREST,
		APIModelName: map[string]string{
			"bedrock": "amazon.nova-lite-v1:0",
		},
		RequiredCredentials: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION"},
		Capabilities: ModelCapabilities{
			SupportsStyles:         false,
			SupportsNegativePrompt: true,
			SupportsSeed:           true,
			MaxPromptLength:        512,
		},
		Pricing: PricingInfo{
			CostPerImage: float64Ptr(0.04),
			FreeTier:     false,
		},
		ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
			if cfg.AWSBedrockAPIKey != "" {
				return NewBedrockRESTClient(cfg.AWSBedrockAPIKey, cfg.AWSRegion)
			}
			ctx := context.Background()
			return NewBedrockSDKClient(ctx, cfg.AWSRegion)
		},
	})
}

// Register adds a model provider to the registry
func (r *Registry) Register(provider *ModelProvider) {
	r.providers[provider.ModelID] = provider
	// Register all aliases
	for _, alias := range provider.Aliases {
		r.aliases[alias] = provider.ModelID
	}
}

// Resolve finds the model provider for a given name or alias
func (r *Registry) Resolve(modelName string) (*ModelProvider, error) {
	// Try exact match first
	if provider, ok := r.providers[modelName]; ok {
		return provider, nil
	}

	// Try alias resolution
	if modelID, ok := r.aliases[modelName]; ok {
		if provider, ok := r.providers[modelID]; ok {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("unknown model: %s", modelName)
}

// CreateClient creates the appropriate client for a model
func (r *Registry) CreateClient(modelName string) (ImageGenerator, error) {
	provider, err := r.Resolve(modelName)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Check credentials
	if err := r.CheckCredentials(provider, cfg); err != nil {
		return nil, err
	}

	// Create client using factory
	return provider.ClientFactory(cfg)
}

// CheckCredentials verifies that required credentials are available
func (r *Registry) CheckCredentials(provider *ModelProvider, cfg *config.Config) error {
	missing := []string{}

	for _, cred := range provider.RequiredCredentials {
		switch cred {
		case "GEMINI_API_KEY":
			if cfg.GeminiAPIKey == "" {
				missing = append(missing, cred)
			}
		case "VERTEX_API_KEY":
			if cfg.VertexAPIKey == "" {
				missing = append(missing, cred)
			}
		case "VERTEX_PROJECT":
			if cfg.VertexProject == "" {
				missing = append(missing, cred)
			}
		case "VERTEX_LOCATION":
			if cfg.VertexLocation == "" {
				missing = append(missing, cred)
			}
		case "AWS_ACCESS_KEY_ID":
			if cfg.AWSAccessKeyID == "" {
				missing = append(missing, cred)
			}
		case "AWS_SECRET_ACCESS_KEY":
			if cfg.AWSSecretAccessKey == "" {
				missing = append(missing, cred)
			}
		case "AWS_REGION":
			if cfg.AWSRegion == "" {
				missing = append(missing, cred)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required credentials for %s: %v", provider.DisplayName, missing)
	}

	return nil
}

// GetAPIModelName returns the correct model name for a specific API
func (r *Registry) GetAPIModelName(modelName string, api string) (string, error) {
	provider, err := r.Resolve(modelName)
	if err != nil {
		return "", err
	}

	if apiName, ok := provider.APIModelName[api]; ok {
		return apiName, nil
	}

	// Default to ModelID if no specific mapping
	return provider.ModelID, nil
}

// ListModels returns all registered models
func (r *Registry) ListModels() []*ModelProvider {
	models := make([]*ModelProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		models = append(models, provider)
	}
	return models
}

// ListAvailableModels returns models that have configured credentials
func (r *Registry) ListAvailableModels() ([]*ModelProvider, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	available := []*ModelProvider{}
	for _, provider := range r.providers {
		if err := r.CheckCredentials(provider, cfg); err == nil {
			available = append(available, provider)
		}
	}
	return available, nil
}

// GenerateWithRegistry is a high-level function that handles the entire generation flow
func GenerateWithRegistry(ctx context.Context, modelName string, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
	registry := GetRegistry()

	// Resolve model
	provider, err := registry.Resolve(modelName)
	if err != nil {
		return nil, err
	}

	// Create appropriate client
	client, err := registry.CreateClient(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", provider.DisplayName, err)
	}
	defer client.Close()

	// Update options with correct API model name
	apiModelName, _ := registry.GetAPIModelName(modelName, provider.API)
	options.Model = apiModelName

	// Generate image
	return client.GenerateImage(ctx, prompt, options)
}