package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chadneal/gimage/internal/config"
	"github.com/chadneal/gimage/internal/generate"
	"github.com/chadneal/gimage/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Helper functions for output
func printVerbose(format string, args ...interface{}) {
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

func printInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func printSuccess(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "✓ "+format+"\n", args...)
}

func formatImageSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

var generateCmd = &cobra.Command{
	Use:   "generate [prompt]",
	Short: "Generate an image from a text prompt using AI",
	Long: `Generate an image from a text prompt using Google Gemini or Vertex AI.

The prompt should describe the image you want to generate. You can optionally
specify style, size, negative prompts, and other parameters.

Examples:
  # List all available models
  gimage generate --list-models

  # Generate with default settings (Gemini 2.5 Flash)
  gimage generate "a sunset over mountains"

  # Generate with specific model
  gimage generate "futuristic city" --model imagen-4

  # Generate with specific style and size
  gimage generate "abstract art" --size 1024x1024 --style photorealistic

  # Use Vertex AI (requires --project for vertex models)
  gimage generate "abstract art" --api vertex --model imagen-4 --project my-project

  # Use negative prompts and seed for reproducibility
  gimage generate "forest scene" --negative "people, buildings" --seed 12345`,
	Args: cobra.ArbitraryArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if --list-models flag is set
		listModels, _ := cmd.Flags().GetBool("list-models")
		if listModels {
			return nil // Skip validation for list-models
		}

		// Require at least one argument (the prompt) if not listing models
		if len(args) == 0 {
			return fmt.Errorf("prompt is required (or use --list-models to see available models)")
		}

		// Validate flags
		size, _ := cmd.Flags().GetString("size")
		if size != "" {
			parts := strings.Split(size, "x")
			if len(parts) != 2 {
				return fmt.Errorf("invalid size format: %s (expected format: WIDTHxHEIGHT, e.g., 1024x1024)", size)
			}
		}
		return nil
	},
	RunE: runGenerate,
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Get flags
	output, _ := cmd.Flags().GetString("output")
	api, _ := cmd.Flags().GetString("api")
	apiKey, _ := cmd.Flags().GetString("api-key")
	// project, _ := cmd.Flags().GetString("project") // TODO: Enable when Vertex is implemented
	// location, _ := cmd.Flags().GetString("location") // TODO: Enable when Vertex is implemented
	model, _ := cmd.Flags().GetString("model")
	size, _ := cmd.Flags().GetString("size")
	style, _ := cmd.Flags().GetString("style")
	negative, _ := cmd.Flags().GetString("negative")
	seed, _ := cmd.Flags().GetInt64("seed")
	listModels, _ := cmd.Flags().GetBool("list-models")

	// Handle --list-models flag
	if listModels {
		return printAvailableModels()
	}

	// Join prompt from args
	prompt := strings.Join(args, " ")

	printVerbose("Generating image with prompt: %s", prompt)

	// Build generate options
	options := models.GenerateOptions{
		Model:          model,
		Size:           size,
		Style:          style,
		NegativePrompt: negative,
		Seed:           seed,
	}

	// Determine which API to use
	// Priority: 1. --api flag, 2. Auto-detect from model, 3. Auto-detect from available credentials
	selectedAPI := api
	if selectedAPI == "" {
		// Auto-detect API from model
		if model != "" {
			detectedAPI, err := generate.DetectAPIFromModel(model)
			if err != nil {
				return fmt.Errorf("invalid model: %w\nUse --list-models to see available models", err)
			}
			selectedAPI = detectedAPI
		} else {
			// Auto-detect from available credentials
			hasGemini := config.HasGeminiCredentials()
			hasVertex := config.HasVertexCredentials()

			if hasGemini && !hasVertex {
				// Only Gemini credentials available
				selectedAPI = "gemini"
				printVerbose("Auto-detected Gemini API (found credentials)")
			} else if hasVertex && !hasGemini {
				// Only Vertex credentials available
				selectedAPI = "vertex"
				printVerbose("Auto-detected Vertex API (found credentials)")
			} else if hasGemini && hasVertex {
				// Both available - check default_api in config, or default to gemini
				cfg, err := config.LoadConfig()
				if err == nil && cfg.DefaultAPI != "" {
					selectedAPI = cfg.DefaultAPI
					printVerbose("Using default API from config: %s", selectedAPI)
				} else {
					selectedAPI = "gemini"
					printVerbose("Both APIs available, defaulting to Gemini")
				}
			} else {
				// No credentials found
				return fmt.Errorf("no API credentials found. Please set up credentials using:\n" +
					"  Gemini:  gimage auth gemini\n" +
					"  Vertex:  gimage auth vertex")
			}
		}
	}

	// Validate model is compatible with selected API
	if model != "" {
		if err := generate.ValidateModelForAPI(model, selectedAPI); err != nil {
			return err
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var generatedImage *models.GeneratedImage
	var err error

	// Generate based on API selection
	if selectedAPI == "gemini" {
		// Use Gemini API (REST implementation)
		key, err := config.GetGeminiAPIKey(apiKey)
		if err != nil {
			return fmt.Errorf("failed to get Gemini API key: %w\nHint: Set GEMINI_API_KEY environment variable or use --api-key flag", err)
		}

		modelName := model
		if modelName == "" {
			modelName = generate.DefaultModel
		}

		printInfo("Generating image using Gemini API (model: %s)...", modelName)

		// Use REST client instead of SDK client
		client, err := generate.NewGeminiRESTClient(key)
		if err != nil {
			return fmt.Errorf("failed to create Gemini client: %w", err)
		}
		defer client.Close()

		generatedImage, err = client.GenerateImage(ctx, prompt, options)
	} else if selectedAPI == "vertex" {
		// Use Vertex AI - check for Express Mode (API key) or Full Mode (service account)

		// Get project and location from flags or environment
		project, _ := cmd.Flags().GetString("project")
		if project == "" {
			project = os.Getenv("VERTEX_PROJECT")
		}
		location, _ := cmd.Flags().GetString("location")
		if location == "" {
			location = os.Getenv("VERTEX_LOCATION")
			if location == "" {
				location = "us-central1"
			}
		}

		// Check for Vertex API key (Express Mode)
		vertexAPIKey, err := config.GetVertexAPIKey(apiKey)
		if err != nil {
			return fmt.Errorf("failed to get Vertex API key: %w", err)
		}

		modelName := model
		if modelName == "" {
			modelName = "imagen-4.0-generate-001" // Default Imagen model
		}

		printInfo("Generating image using Vertex AI Imagen (model: %s)...", modelName)

		// Choose client based on authentication mode
		if vertexAPIKey != "" {
			// Express Mode - Use REST client with API key
			printVerbose("Using Vertex AI Express Mode (API key authentication)")

			// Load project and location from config if not provided
			if project == "" {
				cfg, err := config.LoadConfig()
				if err == nil && cfg.VertexProject != "" {
					project = cfg.VertexProject
				}
			}
			if location == "" {
				cfg, err := config.LoadConfig()
				if err == nil && cfg.VertexLocation != "" {
					location = cfg.VertexLocation
				} else {
					location = "us-central1"
				}
			}

			client, err := generate.NewVertexRESTClient(vertexAPIKey, project, location)
			if err != nil {
				return fmt.Errorf("failed to create Vertex AI REST client: %w", err)
			}
			defer client.Close()

			generatedImage, err = client.GenerateImage(ctx, prompt, options)
		} else {
			// Full Mode - Use SDK client with service account
			printVerbose("Using Vertex AI Full Mode (service account authentication)")

			client, err := generate.NewVertexSDKClient(ctx, project, location)
			if err != nil {
				return fmt.Errorf("failed to create Vertex AI client: %w", err)
			}
			defer client.Close()

			generatedImage, err = client.GenerateImage(ctx, prompt, options)
		}
	} else {
		return fmt.Errorf("invalid API: %s (must be 'gemini' or 'vertex')", selectedAPI)
	}

	if err != nil {
		return fmt.Errorf("failed to generate image: %w", err)
	}

	// Determine output path
	if output == "" {
		output = generate.GenerateOutputPath(generatedImage.Format)
	}

	// Save image
	printInfo("Saving image to: %s", output)
	if err := generate.SaveImage(generatedImage, output); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	// Print success
	printSuccess("Image generated successfully!")
	printInfo("  File: %s", output)
	printInfo("  Size: %s", formatImageSize(int64(len(generatedImage.Data))))
	printInfo("  Dimensions: %dx%d", generatedImage.Width, generatedImage.Height)

	return nil
}

// printAvailableModels displays all available models in a formatted table
func printAvailableModels() error {
	printInfo("Available Models:\n")

	// Group by API
	geminiModels := generate.ListModelsByAPI("gemini")
	vertexModels := generate.ListModelsByAPI("vertex")

	// Print Gemini models
	printSuccess("Gemini API (Free Tier Available):")
	for _, m := range geminiModels {
		defaultMarker := ""
		if m.Name == generate.DefaultModel {
			defaultMarker = " (Default)"
		}
		printInfo("  %-35s %s%s", m.Name, m.DisplayName, defaultMarker)
		printVerbose("                                      %s, up to %s", m.Description, m.MaxSize)
		fmt.Println()
	}

	// Print Vertex models
	printSuccess("\nVertex AI (Paid - Requires GCP):")
	for _, m := range vertexModels {
		premiumMarker := ""
		if m.Quality == "premium" {
			premiumMarker = " ★ Premium"
		}
		printInfo("  %-35s %s%s", m.Name, m.DisplayName, premiumMarker)
		printVerbose("                                      %s, up to %s", m.Description, m.MaxSize)
		fmt.Println()
	}

	printInfo("\nUsage:")
	printInfo("  gimage generate \"your prompt\" --model <model-name>")
	printInfo("\nExamples:")
	printInfo("  gimage generate \"sunset\" --model gemini-2.5-flash-image")
	printInfo("  gimage generate \"abstract art\" --model imagen-4 --project my-gcp-project")

	return nil
}

func init() {
	generateCmd.Flags().StringP("output", "o", "", "Output file path (default: generated_<timestamp>.png)")
	generateCmd.Flags().String("api", "", "API to use: gemini or vertex (auto-detected from model if not specified)")
	generateCmd.Flags().String("api-key", "", "Gemini API key (or use GEMINI_API_KEY env var)")
	generateCmd.Flags().String("project", "", "Vertex AI project ID (or use GIMAGE_VERTEX_PROJECT env var)")
	generateCmd.Flags().String("location", "us-central1", "Vertex AI location")
	generateCmd.Flags().String("model", "", fmt.Sprintf("Model to use (default: %s). Use --list-models to see all", generate.DefaultModel))
	generateCmd.Flags().Bool("list-models", false, "List all available models and exit")
	generateCmd.Flags().String("size", "1024x1024", "Image size (e.g., 1024x1024, 512x512)")
	generateCmd.Flags().String("style", "", "Image style: photorealistic, artistic, anime")
	generateCmd.Flags().String("negative", "", "Negative prompt to avoid certain features")
	generateCmd.Flags().Int64("seed", 0, "Random seed for reproducibility (0 for random)")

	// Bind to viper for config file support
	viper.BindPFlag("generate.api", generateCmd.Flags().Lookup("api"))
	viper.BindPFlag("generate.model", generateCmd.Flags().Lookup("model"))
	viper.BindPFlag("generate.size", generateCmd.Flags().Lookup("size"))

	rootCmd.AddCommand(generateCmd)
}
