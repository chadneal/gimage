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
		Description: "Generate an AI image from a text prompt using Gemini or Vertex AI. Supports multiple models (Gemini 2.5 Flash, Imagen 3, Imagen 4), various sizes up to 2048x2048, and style controls (photorealistic, artistic, anime). Can use negative prompts to exclude unwanted elements and seeds for reproducible generation. IMPORTANT: Always specify an output path (e.g., ~/Desktop/image.png or ~/Documents/image.png) to ensure the file is saved to an accessible location. The tool will automatically try the current directory first, then fall back to the home directory if needed.",
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
					"description": "Image dimensions. Default is 1024x1024. Larger sizes available with Vertex AI.",
					"default":     "1024x1024",
				},
				"model": map[string]interface{}{
					"type": "string",
					"enum": []string{
						"gemini-2.5-flash-image",
						"gemini-2.0-flash-preview-image-generation",
						"imagen-3.0-generate-002",
						"imagen-4",
					},
					"description": "AI model to use. gemini-2.5-flash-image is default and recommended. imagen-4 offers highest quality but requires Vertex AI.",
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

			result := map[string]interface{}{
				"success":     true,
				"output_path": absOutput,
				"size":        size,
				"model":       modelName,
				"prompt":      prompt,
			}

			// Add warning if we had to fall back to a different location
			if pathWarning != "" {
				result["warning"] = pathWarning
			}

			return result, nil
		},
	}

	server.RegisterTool(tool)
}
