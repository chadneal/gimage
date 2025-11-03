package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/generate"
	"github.com/apresai/gimage/internal/mcp"
	"github.com/apresai/gimage/pkg/models"
)

// RegisterGenerateImageTool registers the generate_image tool
func RegisterGenerateImageTool(server *mcp.MCPServer) {
	tool := mcp.Tool{
		Name:        "generate_image",
		Description: "Generate an AI image from a text prompt using Gemini, Vertex AI, or AWS Bedrock. Quick start: generate_image(prompt='sunset over mountains', output='~/Desktop/sunset.png') uses the default free model (Gemini 2.5 Flash, 1024x1024). For higher quality, use model='imagen-4' (paid, requires Vertex AI). Supports various sizes up to 2048x2048, style controls (photorealistic, artistic, anime), negative prompts to exclude unwanted elements, and seeds for reproducible generation. IMPORTANT: Always specify an output path (e.g., ~/Desktop/image.png or ~/Documents/image.png) to ensure the file is saved to an accessible location.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: false, // Creates new files but doesn't modify existing ones
			IdempotentHint:  false, // Each call generates a different image
			ReadOnlyHint:    false, // Writes files to disk
		},
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Text description of the image to generate. Be specific and descriptive for best results.",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output file path. RECOMMENDED: Always specify a path like ~/Desktop/image.png or ~/Documents/image.png. If not provided, will try current directory first, then fall back to home directory. Supports tilde (~) expansion for home directory.",
				},
				"size": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"256x256", "512x512", "1024x1024", "1024x1792", "1792x1024", "2048x2048"},
					"description": "Image dimensions (WIDTHxHEIGHT). Default: 1024x1024. Gemini supports up to 1024x1024. Larger sizes (1792x1024, 2048x2048) require Vertex AI with imagen-4. Examples: '1024x1024' (square), '1792x1024' (16:9 landscape), '1024x1792' (9:16 portrait), '2048x2048' (ultra HD).",
					"default":     "1024x1024",
				},
				"model": map[string]interface{}{
					"type": "string",
					"enum": []string{
						"gemini-2.5-flash-image",
						"gemini-2.0-flash-preview-image-generation",
						"imagen-3.0-generate-002",
						"imagen-4",
						"gemini",
						"gemini-flash",
						"imagen",
						"nova-canvas",
						"amazon.nova-canvas-v1:0",
					},
					"description": "AI model to use. Supports exact names or aliases. Common aliases: 'gemini' or 'gemini-flash' for gemini-2.5-flash-image (default, FREE up to 1500/day, supports up to 1024x1024), 'imagen' or 'imagen-4' for imagen-4.0-generate-001 (paid $0.02-0.04/image, highest quality, supports up to 2048x2048), 'nova-canvas' for amazon.nova-canvas-v1:0 (paid $0.04-0.08/image, supports up to 1408x1408). Invalid model names automatically fall back to default. Examples: 'gemini' (quick iterations), 'imagen-4' (final high-quality output).",
					"default":     "gemini-2.5-flash-image",
				},
				"style": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"photorealistic", "artistic", "anime"},
					"description": "Image style. Affects rendering approach. Optional.",
				},
				"negative": map[string]interface{}{
					"type":        "string",
					"description": "Negative prompt - describe what you DON'T want in the image (e.g., 'people, buildings, modern objects')",
				},
				"seed": map[string]interface{}{
					"type":        "integer",
					"description": "Random seed for reproducible generation. Use the same seed to get the same image.",
				},
			},
			"required": []string{"prompt"},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			// Extract and validate prompt
			prompt, ok := args["prompt"].(string)
			if !ok || prompt == "" {
				return nil, fmt.Errorf("prompt is required and must be a non-empty string")
			}

			// Extract optional parameters
			outputArg, _ := args["output"].(string)

			// Validate and fix output path BEFORE generating image
			// This avoids wasting API calls if the path is not writable
			defaultFilename := fmt.Sprintf("generated_%d.png", time.Now().Unix())
			pathResult, pathErr := ValidateAndFixOutputPath(outputArg, defaultFilename)
			if pathErr != nil {
				return nil, fmt.Errorf("output path validation failed: %w\n\nTIP: Try specifying an explicit output path like ~/Desktop/image.png or ~/Documents/image.png", pathErr)
			}
			output := pathResult.Path

			// Include warning in response if we had to fall back to a different location
			var pathWarning string
			if pathResult.Warning != "" {
				pathWarning = pathResult.Warning
			}

			size, _ := args["size"].(string)
			if size == "" {
				size = "1024x1024"
			}

			modelName, _ := args["model"].(string)
			if modelName == "" {
				modelName = "gemini-2.5-flash-image"
			}

			// Resolve model aliases to exact names (e.g., "gemini" -> "gemini-2.5-flash-image")
			modelName = generate.ResolveModelName(modelName)

			// Validate model exists, fallback to default if not
			_, modelErr := generate.GetModelInfo(modelName)
			if modelErr != nil {
				// Model not found, fallback to default free model
				modelName = "gemini-2.5-flash-image"
			}

			style, _ := args["style"].(string)
			negative, _ := args["negative"].(string)

			var seed int64
			if seedVal, ok := args["seed"].(float64); ok {
				seed = int64(seedVal)
			}

			// Create generate options
			opts := models.GenerateOptions{
				Model:          modelName,
				Size:           size,
				Style:          style,
				NegativePrompt: negative,
				Seed:           seed,
			}

			// Determine which backend to use based on model
			selectedAPI := "gemini" // default
			if isVertexModel(modelName) {
				selectedAPI = "vertex"
			}

			// Create context for API calls
			ctx := context.Background()

			// Generate based on backend
			var generatedImage *models.GeneratedImage
			var err error

			if selectedAPI == "gemini" {
				// Use Gemini REST client
				apiKey, apiErr := config.GetGeminiAPIKey("")
				if apiErr != nil {
					return nil, fmt.Errorf("Gemini API key not configured: %w\nPlease run: gimage auth gemini", apiErr)
				}

				client, err := generate.NewGeminiRESTClient(apiKey)
				if err != nil {
					return nil, fmt.Errorf("failed to create Gemini client: %w", err)
				}
				defer client.Close()

				generatedImage, err = client.GenerateImage(ctx, prompt, opts)
				if err != nil {
					return nil, fmt.Errorf("image generation failed: %w", err)
				}
			} else {
				// Use Vertex AI
				// Load config to get project and location
				cfg, err := config.LoadConfig()
				if err != nil {
					return nil, fmt.Errorf("failed to load config: %w", err)
				}

				project := cfg.VertexProject
				location := cfg.VertexLocation
				if location == "" {
					location = "us-central1"
				}

				// Check for Express Mode (API key) first
				vertexAPIKey, _ := config.GetVertexAPIKey("")

				if vertexAPIKey != "" {
					// Express Mode - Use REST client
					client, err := generate.NewVertexRESTClient(vertexAPIKey, project, location)
					if err != nil {
						return nil, fmt.Errorf("failed to create Vertex AI REST client: %w", err)
					}
					defer client.Close()

					generatedImage, err = client.GenerateImage(ctx, prompt, opts)
					if err != nil {
						return nil, fmt.Errorf("image generation failed: %w", err)
					}
				} else {
					// Full Mode - Use SDK client
					client, err := generate.NewVertexSDKClient(ctx, project, location)
					if err != nil {
						return nil, fmt.Errorf("failed to create Vertex AI SDK client: %w\nPlease run: gimage auth vertex", err)
					}
					defer client.Close()

					generatedImage, err = client.GenerateImage(ctx, prompt, opts)
					if err != nil {
						return nil, fmt.Errorf("image generation failed: %w", err)
					}
				}
			}

			// Save the generated image
			if err := generate.SaveImage(generatedImage, output); err != nil {
				return nil, fmt.Errorf("failed to save image: %w", err)
			}

			// Get absolute output path
			absOutput, err := filepath.Abs(output)
			if err != nil {
				absOutput = output
			}

			// Get model info for cost tracking and announcement
			modelInfo, _ := generate.GetModelInfo(modelName)
			var modelDisplayName string
			var pricingInfo string
			var estimatedCost float64
			var tokensUsed int
			var costExplanation string

			if modelInfo != nil {
				modelDisplayName = modelInfo.DisplayName
				pricingInfo = generate.FormatPricingDisplay(modelInfo)
				estimatedCost, tokensUsed, costExplanation = generate.GetEstimatedCost(modelInfo, size, 1)
			} else {
				modelDisplayName = modelName
				pricingInfo = "Unknown"
			}

			// Build result with comprehensive information
			result := map[string]interface{}{
				"success":       true,
				"output_path":   absOutput,
				"size":          size,
				"model":         modelName,
				"model_display": modelDisplayName,
				"api":           selectedAPI,
				"pricing":       pricingInfo,
				"prompt":        prompt,
			}

			// Add cost and token information
			if tokensUsed > 0 {
				result["tokens_used"] = tokensUsed
			}
			if estimatedCost > 0 {
				result["estimated_cost"] = estimatedCost
				result["cost_explanation"] = costExplanation
			} else if modelInfo != nil && modelInfo.Pricing.FreeTier {
				result["estimated_cost"] = 0
				result["cost_explanation"] = "FREE (within daily limit)"
			}

			// Create user-friendly message
			msg := fmt.Sprintf("Generated using %s (%s). ", modelDisplayName, pricingInfo)
			if tokensUsed > 0 {
				msg += fmt.Sprintf("Used ~%d tokens. ", tokensUsed)
			}
			if estimatedCost > 0 {
				msg += fmt.Sprintf("Cost: $%.4f", estimatedCost)
			} else if modelInfo != nil && modelInfo.Pricing.FreeTier {
				msg += "Cost: FREE"
			}
			result["message"] = msg

			// Add warning if we had to fall back to a different location
			if pathWarning != "" {
				result["warning"] = pathWarning
			}

			return result, nil
		},
	}

	server.RegisterTool(tool)
}
