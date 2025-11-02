package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/generate"
	"github.com/apresai/gimage/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
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

func printWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "⚠ "+format+"\n", args...)
}

func greenYes() string {
	return "✅"
}

func redNo() string {
	return "❌"
}

// padRight pads a string to the right, accounting for ANSI color codes
// which don't contribute to visible width
func padRight(s string, width int) string {
	// Count ANSI escape sequences (they don't contribute to visible width)
	visibleLen := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			visibleLen++
		}
	}

	padding := width - visibleLen
	if padding <= 0 {
		return s
	}

	return s + strings.Repeat(" ", padding)
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
	Long: `Generate an image from a text prompt using Google Gemini, Vertex AI, or AWS Bedrock.

The prompt should describe the image you want to generate. You can optionally
specify style, size, negative prompts, and other parameters.

Examples:
  # List all available models
  gimage generate --list-models

  # Generate with default settings (Gemini 2.5 Flash)
  gimage generate "a sunset over mountains"

  # Generate with specific model (auto-detects API)
  gimage generate "futuristic city" --model imagen-4

  # Generate with AWS Bedrock Nova Canvas (auto-detects bedrock API)
  gimage generate "futuristic city" --model nova-canvas

  # Generate with specific style and size
  gimage generate "abstract art" --size 1024x1024 --style photorealistic

  # Override API selection (rarely needed)
  gimage generate "abstract art" --api vertex --model imagen-4

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
			hasBedrock := config.HasBedrockCredentials()

			// Count available credentials
			availableCount := 0
			if hasGemini {
				availableCount++
			}
			if hasVertex {
				availableCount++
			}
			if hasBedrock {
				availableCount++
			}

			if availableCount == 0 {
				// No credentials found
				return fmt.Errorf("no API credentials found. Please set up credentials using:\n" +
					"  Gemini:  gimage auth gemini\n" +
					"  Vertex:  gimage auth vertex\n" +
					"  Bedrock: gimage auth bedrock")
			} else if availableCount == 1 {
				// Only one API available
				if hasGemini {
					selectedAPI = "gemini"
					printVerbose("Auto-detected Gemini API (found credentials)")
				} else if hasVertex {
					selectedAPI = "vertex"
					printVerbose("Auto-detected Vertex API (found credentials)")
				} else {
					selectedAPI = "bedrock"
					printVerbose("Auto-detected AWS Bedrock API (found credentials)")
				}
			} else {
				// Multiple APIs available - check default_api in config, or default to gemini
				cfg, err := config.LoadConfig()
				if err == nil && cfg.DefaultAPI != "" {
					selectedAPI = cfg.DefaultAPI
					printVerbose("Using default API from config: %s", selectedAPI)
				} else {
					selectedAPI = "gemini"
					printVerbose("Multiple APIs available, defaulting to Gemini")
				}
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

		// Get model info and announce selection
		modelInfo, _ := generate.GetModelInfo(modelName)
		if modelInfo != nil {
			printInfo("Using: %s (%s API)", modelInfo.DisplayName, modelInfo.API)
			printInfo("Pricing: %s", generate.FormatPricingDisplay(modelInfo))

			// Calculate estimated cost
			cost, tokens, explanation := generate.GetEstimatedCost(modelInfo, size, 1)
			if tokens > 0 {
				printVerbose("Estimated: %s", explanation)
			}

			// Warn if expensive (cost > $0.05)
			if cost > 0.05 {
				fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", modelInfo.DisplayName, *modelInfo.Pricing.CostPerImage)
			}
		}

		printInfo("Generating image...")

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

		// Get model info and announce selection
		modelInfo, _ := generate.GetModelInfo(modelName)
		if modelInfo != nil {
			printInfo("Using: %s (%s API)", modelInfo.DisplayName, modelInfo.API)
			printInfo("Pricing: %s", generate.FormatPricingDisplay(modelInfo))

			// Calculate estimated cost
			cost, tokens, explanation := generate.GetEstimatedCost(modelInfo, size, 1)
			if tokens > 0 {
				printVerbose("Estimated: %s", explanation)
			} else {
				printVerbose("Estimated: %s", explanation)
			}

			// Warn if expensive (cost > $0.05)
			if cost > 0.05 {
				fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", modelInfo.DisplayName, *modelInfo.Pricing.CostPerImage)
			}
		}

		printInfo("Generating image...")

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
	} else if selectedAPI == "bedrock" {
		// Use AWS Bedrock API - choose between REST (bearer token) or SDK (IAM/keys)
		region := config.GetAWSRegion("")
		printVerbose("Using AWS Bedrock region: %s", region)

		modelName := model
		if modelName == "" {
			modelName = generate.ModelNovaCanvas
		}

		printVerbose("Using model: %s", modelName)

		// Resolve model alias and show info
		modelInfo, err := generate.GetModelInfo(modelName)
		if err != nil {
			return fmt.Errorf("unknown model: %s", modelName)
		}

		// Use the resolved full model ID
		resolvedModelID := modelInfo.Name
		printVerbose("Resolved model ID: %s", resolvedModelID)

		printVerbose("Model: %s (%s)", modelInfo.DisplayName, modelInfo.Quality)
		printVerbose("Max resolution: %s", modelInfo.Pricing.MaxResolution)

		// Show pricing info
		if !modelInfo.Free {
			cost, _, explanation := generate.GetEstimatedCost(modelInfo, size, 1)
			printVerbose("Estimated cost: %s", explanation)

			// Warn if expensive (cost > $0.05)
			if cost > 0.05 {
				fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", modelInfo.DisplayName, *modelInfo.Pricing.CostPerImage)
			}
		}

		// Update options to use resolved model ID (not alias)
		bedrockOptions := options
		bedrockOptions.Model = resolvedModelID

		// Determine which authentication method to use
		// Priority: Bearer token (REST) > AWS SDK (keys/profile/IAM)
		cfg, _ := config.LoadConfig()
		bearerToken := os.Getenv("AWS_BEARER_TOKEN_BEDROCK")
		if bearerToken == "" && cfg != nil {
			bearerToken = cfg.AWSBedrockAPIKey
		}

		if bearerToken != "" {
			// Use REST client with bearer token
			printVerbose("Using Bedrock REST API with bearer token authentication")
			printInfo("Generating image with AWS Bedrock (REST API)...")

			client, err := generate.NewBedrockRESTClient(bearerToken, region)
			if err != nil {
				return fmt.Errorf("failed to create AWS Bedrock REST client: %w", err)
			}
			defer client.Close()

			generatedImage, err = client.GenerateImage(ctx, prompt, bedrockOptions)
		} else {
			// Use SDK client with IAM/keys/profile
			printVerbose("Using Bedrock SDK with IAM/keys/profile authentication")
			printInfo("Generating image with AWS Bedrock (SDK)...")

			client, err := generate.NewBedrockSDKClient(ctx, region)
			if err != nil {
				return fmt.Errorf("failed to create AWS Bedrock SDK client: %w", err)
			}
			defer client.Close()

			generatedImage, err = client.GenerateImage(ctx, prompt, bedrockOptions)
		}
	} else {
		return fmt.Errorf("invalid API: %s (must be 'gemini', 'vertex', or 'bedrock')", selectedAPI)
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

	// Print success with cost tracking
	printSuccess("Image generated successfully!")
	printInfo("  File: %s", output)
	printInfo("  Size: %s", formatImageSize(int64(len(generatedImage.Data))))
	printInfo("  Dimensions: %dx%d", generatedImage.Width, generatedImage.Height)

	// Log cost and token usage
	modelNameUsed := model
	if modelNameUsed == "" {
		if selectedAPI == "gemini" {
			modelNameUsed = generate.DefaultModel
		} else {
			modelNameUsed = "imagen-4.0-generate-001"
		}
	}

	if modelInfo, err := generate.GetModelInfo(modelNameUsed); err == nil {
		cost, tokens, _ := generate.GetEstimatedCost(modelInfo, size, 1)
		if tokens > 0 {
			printInfo("  Tokens used: ~%d tokens", tokens)
		}
		if cost > 0 {
			printInfo("  Cost: $%.4f", cost)
		} else if modelInfo.Pricing.FreeTier {
			printInfo("  Cost: FREE (within daily limit)")
		}
	}

	return nil
}

// printAvailableModels displays all available models in a formatted table
func printAvailableModels() error {
	// Check which APIs have credentials
	hasGemini := config.HasGeminiCredentials()
	hasVertex := config.HasVertexCredentials()
	hasBedrock := config.HasBedrockCredentials()
	hasAnyAuth := hasGemini || hasVertex || hasBedrock

	// ASCII art header
	printInfo("\n╔═══════════════════════════════════════════════════════════════════════════════╗")
	printInfo("║                                                                               ║")
	printInfo("║     █████╗ ██╗   ██╗ █████╗ ██╗██╗      █████╗ ██████╗ ██╗     ███████╗     ║")
	printInfo("║    ██╔══██╗██║   ██║██╔══██╗██║██║     ██╔══██╗██╔══██╗██║     ██╔════╝     ║")
	printInfo("║    ███████║██║   ██║███████║██║██║     ███████║██████╔╝██║     █████╗       ║")
	printInfo("║    ██╔══██║╚██╗ ██╔╝██╔══██║██║██║     ██╔══██║██╔══██╗██║     ██╔══╝       ║")
	printInfo("║    ██║  ██║ ╚████╔╝ ██║  ██║██║███████╗██║  ██║██████╔╝███████╗███████╗     ║")
	printInfo("║    ╚═╝  ╚═╝  ╚═══╝  ╚═╝  ╚═╝╚═╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝     ║")
	printInfo("║                                                                               ║")
	printInfo("║               🎨  AI Image Generation Models  🎨                              ║")
	printInfo("║                                                                               ║")
	printInfo("╚═══════════════════════════════════════════════════════════════════════════════╝\n")

	if !hasAnyAuth {
		printWarning("⚠️  No API credentials configured. Set up authentication to use these models:\n")
	}

	// Print Gemini models - ALWAYS show, indicate auth status
	geminiModels := generate.ListModelsByAPI("gemini")
	printInfo("┌─────────────────────────────────────────────────────────────────────────────────┐")
	if hasGemini {
		printSuccess("│ ✓ Gemini API (AUTHENTICATED - Free Tier Available)                             │")
	} else {
		printWarning("│ ○ Gemini API (NOT AUTHENTICATED - Setup: gimage auth gemini)                   │")
	}
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	printInfo("│ Models:                                                                         │")
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	for _, m := range geminiModels {
		pricing := generate.FormatPricingDisplay(&m)
		priorityMark := fmt.Sprintf("%d", m.Priority)
		if m.Name == generate.DefaultModel {
			priorityMark = fmt.Sprintf("%d ⭐", m.Priority)
		}
		authMark := greenYes()
		if !hasGemini {
			authMark = redNo()
		}
		// Display alias if available, otherwise full name
		displayName := m.Name
		if alias := generate.GetPreferredAlias(m.Name); alias != "" {
			displayName = fmt.Sprintf("%s (%s)", alias, m.Name)
		}
		// Print model name and display name on first line (73 chars for model name)
		paddedName := padRight(displayName, 73)
		printInfo("│ %s  %s │", authMark, paddedName)
		// Print details on second line (indented)
		printInfo("│     Priority: %-3s  Pricing: %-55s │", priorityMark, pricing)
		if viper.GetBool("verbose") {
			printVerbose("│     %s", m.Description)
			if m.Pricing.TokensPerImage != nil {
				printVerbose("│     Tokens per image: ~%d", *m.Pricing.TokensPerImage)
			}
		}
	}
	printInfo("└─────────────────────────────────────────────────────────────────────────────────┘\n")

	// Print Vertex models - ALWAYS show, indicate auth status
	vertexModels := generate.ListModelsByAPI("vertex")
	printInfo("┌─────────────────────────────────────────────────────────────────────────────────┐")
	if hasVertex {
		printSuccess("│ ✓ Vertex AI (AUTHENTICATED - Paid, Requires GCP)                               │")
	} else {
		printWarning("│ ○ Vertex AI (NOT AUTHENTICATED - Setup: gimage auth vertex)                    │")
	}
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	printInfo("│ Models:                                                                         │")
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	for _, m := range vertexModels {
		pricing := generate.FormatPricingDisplay(&m)
		priorityMark := fmt.Sprintf("%d", m.Priority)
		if m.Quality == "premium" {
			priorityMark = fmt.Sprintf("%d ★", m.Priority)
		}
		authMark := greenYes()
		if !hasVertex {
			authMark = redNo()
		}
		// Display alias if available, otherwise full name
		displayName := m.Name
		if alias := generate.GetPreferredAlias(m.Name); alias != "" {
			displayName = fmt.Sprintf("%s (%s)", alias, m.Name)
		}
		// Print model name and display name on first line (73 chars for model name)
		paddedName := padRight(displayName, 73)
		printInfo("│ %s  %s │", authMark, paddedName)
		// Print details on second line (indented)
		printInfo("│     Priority: %-3s  Pricing: %-55s │", priorityMark, pricing)
		if viper.GetBool("verbose") {
			printVerbose("│     %s", m.Description)
		}
	}
	printInfo("└─────────────────────────────────────────────────────────────────────────────────┘\n")

	// Print Bedrock models - ALWAYS show, indicate auth status
	bedrockModels := generate.ListModelsByAPI("bedrock")
	printInfo("┌─────────────────────────────────────────────────────────────────────────────────┐")
	if hasBedrock {
		printSuccess("│ ✓ AWS Bedrock (AUTHENTICATED - Paid, Requires AWS)                             │")
	} else {
		printWarning("│ ○ AWS Bedrock (NOT AUTHENTICATED - Setup: gimage auth bedrock)                 │")
	}
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	printInfo("│ Models:                                                                         │")
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	for _, m := range bedrockModels {
		pricing := generate.FormatPricingDisplay(&m)
		priorityMark := fmt.Sprintf("%d", m.Priority)
		if m.Quality == "premium" {
			priorityMark = fmt.Sprintf("%d ★", m.Priority)
		}
		authMark := greenYes()
		if !hasBedrock {
			authMark = redNo()
		}
		// Display alias if available, otherwise full name
		displayName := m.Name
		if alias := generate.GetPreferredAlias(m.Name); alias != "" {
			displayName = fmt.Sprintf("%s (%s)", alias, m.Name)
		}
		// Print model name and display name on first line (73 chars for model name)
		paddedName := padRight(displayName, 73)
		printInfo("│ %s  %s │", authMark, paddedName)
		// Print details on second line (indented)
		printInfo("│     Priority: %-3s  Pricing: %-55s │", priorityMark, pricing)
		if viper.GetBool("verbose") {
			printVerbose("│     %s", m.Description)
		}
	}
	printInfo("└─────────────────────────────────────────────────────────────────────────────────┘\n")

	printInfo("╔═══════════════════════════════════════════════════════════════════════════════╗")
	printInfo("║                                   LEGEND                                      ║")
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  %s  = Authenticated (ready to use)                                          ║", greenYes())
	printInfo("║  %s  = Not authenticated (run setup command)                                 ║", redNo())
	printInfo("║  ⭐  = Default model (auto-selected)                                          ║")
	printInfo("║  ★  = Premium quality                                                        ║")
	printInfo("║  Lower priority number = higher priority for auto-selection                 ║")
	printInfo("╚═══════════════════════════════════════════════════════════════════════════════╝")

	printInfo("\n╔═══════════════════════════════════════════════════════════════════════════════╗")
	printInfo("║                                 QUICK START                                   ║")
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  Usage: gimage generate \"your prompt\" --model <model-name>                   ║")
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  Recommended Models:                                                          ║")
	if hasGemini {
		printInfo("║    ✓ Free users:  gemini (500/day FREE)                                       ║")
	} else {
		printInfo("║    ○ Free users:  gemini (setup: gimage auth gemini)                          ║")
	}
	if hasVertex {
		printInfo("║    ✓ Paid users:  imagen-4-fast ($0.02/image, fastest paid)                  ║")
	} else {
		printInfo("║    ○ Paid users:  imagen-4-fast (setup: gimage auth vertex)                  ║")
	}
	if hasBedrock {
		printInfo("║    ✓ AWS users:   nova-canvas ($0.04/image standard, $0.08 premium)          ║")
	} else {
		printInfo("║    ○ AWS users:   nova-canvas (setup: gimage auth bedrock)                   ║")
	}
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  Examples:                                                                    ║")
	if hasGemini {
		printInfo("║    gimage generate \"sunset\" --model gemini                                   ║")
	}
	if hasVertex {
		printInfo("║    gimage generate \"abstract art\" --model imagen-4-fast                      ║")
	}
	if hasBedrock {
		printInfo("║    gimage generate \"landscape\" --api bedrock                                 ║")
		printInfo("║    gimage generate \"portrait\" --model nova-canvas                            ║")
	}
	printInfo("╚═══════════════════════════════════════════════════════════════════════════════╝")

	if !hasAnyAuth {
		printInfo("\n╔═══════════════════════════════════════════════════════════════════════════════╗")
		printInfo("║                            GET STARTED                                        ║")
		printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
		printInfo("║  To get started, authenticate with at least one API:                         ║")
		printInfo("║                                                                               ║")
		printInfo("║    gimage auth gemini   # Fastest, has free tier                             ║")
		printInfo("║    gimage auth vertex   # Highest quality (paid)                             ║")
		printInfo("║    gimage auth bedrock  # AWS integration (paid)                             ║")
		printInfo("╚═══════════════════════════════════════════════════════════════════════════════╝")
	}

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
