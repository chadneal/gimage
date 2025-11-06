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

You can provide the prompt as a positional argument (recommended for quick use)
or use the --prompt flag for explicit clarity.

Examples:
  # List all available models
  gimage generate --list-models

  # Generate with default settings - positional prompt (most common)
  gimage generate "a sunset over mountains"

  # Or use --prompt flag explicitly
  gimage generate --prompt "a sunset over mountains"

  # Generate with specific model (auto-detects API)
  gimage generate "futuristic city" --model imagen-4

  # Generate with AWS Bedrock Nova Canvas
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

		// Check if --list-providers flag is set
		listProviders, _ := cmd.Flags().GetBool("list-providers")
		if listProviders {
			return nil // Skip validation for list-providers
		}

		// Require either positional prompt or --prompt flag
		prompt, _ := cmd.Flags().GetString("prompt")
		if len(args) == 0 && prompt == "" {
			return fmt.Errorf("prompt is required (provide as argument or use --prompt flag)\nExamples:\n  gimage generate \"your prompt here\"\n  gimage generate --prompt \"your prompt here\"\n  gimage generate --list-models")
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
	providerID, _ := cmd.Flags().GetString("provider")
	api, _ := cmd.Flags().GetString("api")
	apiKey, _ := cmd.Flags().GetString("api-key")
	// project, _ := cmd.Flags().GetString("project") // TODO: Enable when Vertex is implemented
	// location, _ := cmd.Flags().GetString("location") // TODO: Enable when Vertex is implemented
	model, _ := cmd.Flags().GetString("model")

	// Resolve model aliases to exact names (e.g., "gemini" -> "gemini-2.5-flash-image")
	originalModel := model
	if model != "" {
		model = generate.ResolveModelName(model)
		if originalModel != model {
			printVerbose("Resolved model '%s' to '%s'", originalModel, model)
		}
	}

	size, _ := cmd.Flags().GetString("size")
	style, _ := cmd.Flags().GetString("style")
	negative, _ := cmd.Flags().GetString("negative")
	seed, _ := cmd.Flags().GetInt64("seed")
	listModels, _ := cmd.Flags().GetBool("list-models")
	listProviders, _ := cmd.Flags().GetBool("list-providers")

	// Handle --list-providers flag
	if listProviders {
		return printAvailableProviders()
	}

	// Handle --list-models flag (legacy)
	if listModels {
		return printAvailableModels()
	}

	// Get prompt - prefer positional argument, fall back to --prompt flag
	var prompt string
	if len(args) > 0 {
		prompt = strings.Join(args, " ")
	} else {
		prompt, _ = cmd.Flags().GetString("prompt")
	}

	printVerbose("Generating image with prompt: %s", prompt)

	// Handle new provider system if --provider is specified
	if providerID != "" {
		return runGenerateWithProvider(cmd, prompt, providerID, output, size, style, negative, seed)
	}

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

		// Get provider info and announce selection
		registry := generate.GetProviderRegistry()
		provider, _ := registry.ResolveProvider(modelName)
		if provider != nil {
			printInfo("Using: %s (%s API)", provider.Name, provider.API)
			pricingDisplay := "Variable"
			if provider.Pricing.FreeTier {
				pricingDisplay = fmt.Sprintf("FREE (%s)", provider.Pricing.FreeTierLimit)
			} else if provider.Pricing.CostPerImage != nil {
				pricingDisplay = fmt.Sprintf("$%.4f/image", *provider.Pricing.CostPerImage)
			}
			printInfo("Pricing: %s", pricingDisplay)

			// Warn if expensive (cost > $0.05)
			if provider.Pricing.CostPerImage != nil && *provider.Pricing.CostPerImage > 0.05 {
				fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", provider.Name, *provider.Pricing.CostPerImage)
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

		// Get provider info and announce selection
		registry := generate.GetProviderRegistry()
		provider, _ := registry.ResolveProvider(modelName)
		if provider != nil {
			printInfo("Using: %s (%s API)", provider.Name, provider.API)
			pricingDisplay := "Variable"
			if provider.Pricing.FreeTier {
				pricingDisplay = fmt.Sprintf("FREE (%s)", provider.Pricing.FreeTierLimit)
			} else if provider.Pricing.CostPerImage != nil {
				pricingDisplay = fmt.Sprintf("$%.4f/image", *provider.Pricing.CostPerImage)
			}
			printInfo("Pricing: %s", pricingDisplay)

			// Warn if expensive (cost > $0.05)
			if provider.Pricing.CostPerImage != nil && *provider.Pricing.CostPerImage > 0.05 {
				fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", provider.Name, *provider.Pricing.CostPerImage)
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

		// Resolve provider alias and show info
		registry := generate.GetProviderRegistry()
		provider, err := registry.ResolveProvider(modelName)
		if err != nil {
			return fmt.Errorf("unknown provider/model: %s", modelName)
		}

		// Use the resolved full model ID
		resolvedModelID := provider.ModelID
		printVerbose("Resolved model ID: %s (Provider: %s)", resolvedModelID, provider.ID)

		printVerbose("Provider: %s", provider.Name)
		printVerbose("Model ID: %s", provider.ModelID)

		// Show pricing info
		if !provider.Pricing.FreeTier && provider.Pricing.CostPerImage != nil {
			cost := *provider.Pricing.CostPerImage
			printVerbose("Cost: $%.4f/image", cost)

			// Warn if expensive (cost > $0.05)
			if cost > 0.05 {
				fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", provider.Name, cost)
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

	// Defensive nil check (should never happen if error handling is correct)
	if generatedImage == nil {
		return fmt.Errorf("internal error: generated image is nil but no error was returned - please report this bug")
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

	registry := generate.GetProviderRegistry()
	if provider, err := registry.ResolveProvider(modelNameUsed); err == nil {
		if provider.Pricing.FreeTier {
			printInfo("  Cost: FREE (within %s)", provider.Pricing.FreeTierLimit)
		} else if provider.Pricing.CostPerImage != nil {
			printInfo("  Cost: $%.4f", *provider.Pricing.CostPerImage)
		}
	}

	return nil
}

// runGenerateWithProvider handles image generation using the new provider system
func runGenerateWithProvider(cmd *cobra.Command, prompt, providerID, output, size, style, negative string, seed int64) error {
	registry := generate.GetProviderRegistry()

	// Resolve provider
	provider, err := registry.ResolveProvider(providerID)
	if err != nil {
		// Show available providers on error
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printAvailableProviders()
		return fmt.Errorf("unknown provider: %s", providerID)
	}

	// Check authentication
	hasAuth, missing, err := registry.CheckAuth(provider)
	if err != nil {
		return fmt.Errorf("failed to check authentication: %w", err)
	}

	if !hasAuth {
		fmt.Fprintf(os.Stderr, "Provider '%s' is not configured.\n", provider.ID)
		fmt.Fprintf(os.Stderr, "Missing credentials: %v\n\n", missing)
		fmt.Fprintf(os.Stderr, "To set up authentication:\n")
		fmt.Fprintf(os.Stderr, "  gimage auth setup %s\n", provider.ID)
		return fmt.Errorf("authentication required")
	}

	// Show provider info
	printInfo("Using provider: %s", provider.Name)
	if provider.Pricing.FreeTier {
		printInfo("Pricing: FREE (%s)", provider.Pricing.FreeTierLimit)
	} else if provider.Pricing.CostPerImage != nil {
		cost := *provider.Pricing.CostPerImage
		printInfo("Pricing: $%.4f per image", cost)

		// Warn if expensive
		if cost > 0.05 {
			fmt.Fprintf(os.Stderr, "⚠️  This will cost $%.4f per image\n", cost)
		}
	}

	// Create client
	printInfo("Creating client for %s...", provider.API)
	client, err := registry.CreateClient(provider.ID)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Prepare options
	options := models.GenerateOptions{
		Model:          provider.ModelID,
		Size:           size,
		Style:          style,
		NegativePrompt: negative,
		Seed:           seed,
	}

	// Generate image
	printInfo("Generating image...")
	ctx := context.Background()

	startTime := time.Now()
	generatedImage, err := client.GenerateImage(ctx, prompt, options)
	if err != nil {
		return fmt.Errorf("failed to generate image: %w", err)
	}

	elapsed := time.Since(startTime)
	printInfo("Generation completed in %.2fs", elapsed.Seconds())

	// Determine output path
	if output == "" {
		output = generate.GenerateOutputPath(generatedImage.Format)
	}

	// Save image
	printInfo("Saving image to: %s", output)
	if err := generate.SaveImage(generatedImage, output); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	// Print success with details
	printSuccess("Image generated successfully!")
	printInfo("  Provider: %s", provider.Name)
	printInfo("  File: %s", output)
	printInfo("  Size: %s", formatImageSize(int64(len(generatedImage.Data))))
	printInfo("  Dimensions: %dx%d", generatedImage.Width, generatedImage.Height)

	// Show cost info
	if provider.Pricing.FreeTier {
		printInfo("  Cost: FREE (within daily limit)")
	} else if provider.Pricing.CostPerImage != nil {
		printInfo("  Cost: $%.4f", *provider.Pricing.CostPerImage)
	}

	return nil
}

// printAvailableProviders prints the list of available providers with auth status
func printAvailableProviders() error {
	registry := generate.GetProviderRegistry()
	statuses := registry.GetAuthStatus()

	fmt.Println("Available Providers:")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Group by API
	geminiProviders := []generate.AuthStatus{}
	vertexProviders := []generate.AuthStatus{}
	bedrockProviders := []generate.AuthStatus{}

	for _, status := range statuses {
		switch status.Provider.API {
		case "gemini":
			geminiProviders = append(geminiProviders, status)
		case "vertex":
			vertexProviders = append(vertexProviders, status)
		case "bedrock":
			bedrockProviders = append(bedrockProviders, status)
		}
	}

	// Print Gemini providers
	if len(geminiProviders) > 0 {
		fmt.Println("Gemini API (Google AI Studio):")
		for _, status := range geminiProviders {
			printProviderStatus(status)
		}
		fmt.Println()
	}

	// Print Vertex providers
	if len(vertexProviders) > 0 {
		fmt.Println("Vertex AI (Google Cloud):")
		for _, status := range vertexProviders {
			printProviderStatus(status)
		}
		fmt.Println()
	}

	// Print Bedrock providers
	if len(bedrockProviders) > 0 {
		fmt.Println("AWS Bedrock:")
		for _, status := range bedrockProviders {
			printProviderStatus(status)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Println("\nUsage:")
	fmt.Println("  gimage generate \"prompt\" --provider <provider-id>")
	fmt.Println("\nExamples:")
	fmt.Println("  gimage generate \"sunset\" --provider gemini/flash-2.5")
	fmt.Println("  gimage generate \"portrait\" --provider vertex/imagen-4")
	fmt.Println("\nTo configure authentication:")
	fmt.Println("  gimage auth setup <provider-id>")

	return nil
}

func printProviderStatus(status generate.AuthStatus) {
	p := status.Provider

	// Status icon
	statusIcon := "✗"
	statusText := "Not configured"
	if status.Configured {
		statusIcon = "✓"
		statusText = "Ready"
	}

	// Pricing
	pricing := "Variable"
	if p.Pricing.FreeTier {
		pricing = fmt.Sprintf("FREE (%s)", p.Pricing.FreeTierLimit)
	} else if p.Pricing.CostPerImage != nil {
		pricing = fmt.Sprintf("$%.4f/image", *p.Pricing.CostPerImage)
	}

	fmt.Printf("  %s %-20s - %-30s [%s] %s\n",
		statusIcon,
		p.ID,
		p.Name,
		pricing,
		statusText,
	)
}

// printAvailableModels displays all available providers in a formatted table
func printAvailableModels() error {
	// Get provider registry and auth status
	registry := generate.GetProviderRegistry()
	statuses := registry.GetAuthStatus()

	// Group providers by API
	geminiProviders := []generate.AuthStatus{}
	vertexProviders := []generate.AuthStatus{}
	bedrockProviders := []generate.AuthStatus{}

	for _, status := range statuses {
		switch status.Provider.API {
		case "gemini":
			geminiProviders = append(geminiProviders, status)
		case "vertex":
			vertexProviders = append(vertexProviders, status)
		case "bedrock":
			bedrockProviders = append(bedrockProviders, status)
		}
	}

	// Check if any auth is configured
	hasAnyAuth := false
	for _, status := range statuses {
		if status.Configured {
			hasAnyAuth = true
			break
		}
	}

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
	printInfo("║               🎨  AI Image Generation Providers  🎨                           ║")
	printInfo("║                                                                               ║")
	printInfo("╚═══════════════════════════════════════════════════════════════════════════════╝\n")

	if !hasAnyAuth {
		printWarning("⚠️  No API credentials configured. Set up authentication to use these providers:\n")
	}

	// Print Gemini providers - ALWAYS show, indicate auth status
	hasGemini := false
	for _, status := range geminiProviders {
		if status.Configured {
			hasGemini = true
			break
		}
	}
	printInfo("┌─────────────────────────────────────────────────────────────────────────────────┐")
	if hasGemini {
		printSuccess("│ ✓ Gemini API (AUTHENTICATED - Free Tier Available)                             │")
	} else {
		printWarning("│ ○ Gemini API (NOT AUTHENTICATED - Setup: gimage auth gemini)                   │")
	}
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	printInfo("│ Providers:                                                                      │")
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	for _, status := range geminiProviders {
		p := status.Provider
		// Format pricing
		pricingDisplay := "Variable"
		if p.Pricing.FreeTier {
			pricingDisplay = fmt.Sprintf("FREE (%s)", p.Pricing.FreeTierLimit)
		} else if p.Pricing.CostPerImage != nil {
			pricingDisplay = fmt.Sprintf("$%.4f/image", *p.Pricing.CostPerImage)
		}

		authMark := greenYes()
		if !status.Configured {
			authMark = redNo()
		}
		// Display provider name (73 chars for name)
		paddedName := padRight(p.Name, 73)
		printInfo("│ %s  %s │", authMark, paddedName)
		// Print details on second line (indented)
		printInfo("│     Provider ID: %-20s  Pricing: %-35s │", p.ID, pricingDisplay)
		if viper.GetBool("verbose") {
			printVerbose("│     %s", p.Description)
			printVerbose("│     Model ID: %s", p.ModelID)
		}
	}
	printInfo("└─────────────────────────────────────────────────────────────────────────────────┘\n")

	// Print Vertex providers - ALWAYS show, indicate auth status
	hasVertex := false
	for _, status := range vertexProviders {
		if status.Configured {
			hasVertex = true
			break
		}
	}
	printInfo("┌─────────────────────────────────────────────────────────────────────────────────┐")
	if hasVertex {
		printSuccess("│ ✓ Vertex AI (AUTHENTICATED - Paid, Requires GCP)                               │")
	} else {
		printWarning("│ ○ Vertex AI (NOT AUTHENTICATED - Setup: gimage auth vertex)                    │")
	}
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	printInfo("│ Providers:                                                                      │")
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	for _, status := range vertexProviders {
		p := status.Provider
		// Format pricing
		pricingDisplay := "Variable"
		if p.Pricing.CostPerImage != nil {
			pricingDisplay = fmt.Sprintf("$%.4f/image", *p.Pricing.CostPerImage)
		}

		authMark := greenYes()
		if !status.Configured {
			authMark = redNo()
		}
		// Display provider name (73 chars for name)
		paddedName := padRight(p.Name, 73)
		printInfo("│ %s  %s │", authMark, paddedName)
		// Print details on second line (indented)
		printInfo("│     Provider ID: %-20s  Pricing: %-35s │", p.ID, pricingDisplay)
		if viper.GetBool("verbose") {
			printVerbose("│     %s", p.Description)
			printVerbose("│     Model ID: %s", p.ModelID)
		}
	}
	printInfo("└─────────────────────────────────────────────────────────────────────────────────┘\n")

	// Print Bedrock providers - ALWAYS show, indicate auth status
	hasBedrock := false
	for _, status := range bedrockProviders {
		if status.Configured {
			hasBedrock = true
			break
		}
	}
	printInfo("┌─────────────────────────────────────────────────────────────────────────────────┐")
	if hasBedrock {
		printSuccess("│ ✓ AWS Bedrock (AUTHENTICATED - Paid, Requires AWS)                             │")
	} else {
		printWarning("│ ○ AWS Bedrock (NOT AUTHENTICATED - Setup: gimage auth bedrock)                 │")
	}
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	printInfo("│ Providers:                                                                      │")
	printInfo("├─────────────────────────────────────────────────────────────────────────────────┤")
	for _, status := range bedrockProviders {
		p := status.Provider
		// Format pricing
		pricingDisplay := "Variable"
		if p.Pricing.CostPerImage != nil {
			pricingDisplay = fmt.Sprintf("$%.4f/image", *p.Pricing.CostPerImage)
		}

		authMark := greenYes()
		if !status.Configured {
			authMark = redNo()
		}
		// Display provider name (73 chars for name)
		paddedName := padRight(p.Name, 73)
		printInfo("│ %s  %s │", authMark, paddedName)
		// Print details on second line (indented)
		printInfo("│     Provider ID: %-20s  Pricing: %-35s │", p.ID, pricingDisplay)
		if viper.GetBool("verbose") {
			printVerbose("│     %s", p.Description)
			printVerbose("│     Model ID: %s", p.ModelID)
		}
	}
	printInfo("└─────────────────────────────────────────────────────────────────────────────────┘\n")

	printInfo("╔═══════════════════════════════════════════════════════════════════════════════╗")
	printInfo("║                                   LEGEND                                      ║")
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  %s  = Authenticated (ready to use)                                          ║", greenYes())
	printInfo("║  %s  = Not authenticated (run setup command)                                 ║", redNo())
	printInfo("╚═══════════════════════════════════════════════════════════════════════════════╝")

	printInfo("\n╔═══════════════════════════════════════════════════════════════════════════════╗")
	printInfo("║                                 QUICK START                                   ║")
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  Usage: gimage generate \"your prompt\" --model <model-name>                   ║")
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  Recommended Providers:                                                       ║")
	if hasGemini {
		printInfo("║    ✓ Free users:  gemini (500/day FREE)                                       ║")
	} else {
		printInfo("║    ○ Free users:  gemini (setup: gimage auth gemini)                          ║")
	}
	if hasVertex {
		printInfo("║    ✓ Paid users:  imagen-4 ($0.04/image, highest quality)                    ║")
	} else {
		printInfo("║    ○ Paid users:  imagen-4 (setup: gimage auth vertex)                       ║")
	}
	if hasBedrock {
		printInfo("║    ✓ AWS users:   nova-canvas ($0.08/image)                                  ║")
	} else {
		printInfo("║    ○ AWS users:   nova-canvas (setup: gimage auth bedrock)                   ║")
	}
	printInfo("╠═══════════════════════════════════════════════════════════════════════════════╣")
	printInfo("║  Examples:                                                                    ║")
	if hasGemini {
		printInfo("║    gimage generate \"sunset\" --model gemini                                   ║")
	}
	if hasVertex {
		printInfo("║    gimage generate \"abstract art\" --model imagen-4                           ║")
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
	generateCmd.Flags().StringP("prompt", "p", "", "Text description of the image to generate (required)")
	generateCmd.Flags().StringP("output", "o", "", "Output file path (default: generated_<timestamp>.png)")
	generateCmd.Flags().String("provider", "", "Provider to use (e.g., gemini/flash-2.5, vertex/imagen-4)")
	generateCmd.Flags().String("api", "", "API to use: gemini or vertex (deprecated, use --provider)")
	generateCmd.Flags().String("api-key", "", "Gemini API key (or use GEMINI_API_KEY env var)")
	generateCmd.Flags().String("project", "", "Vertex AI project ID (or use GIMAGE_VERTEX_PROJECT env var)")
	generateCmd.Flags().String("location", "us-central1", "Vertex AI location")
	generateCmd.Flags().String("model", "", fmt.Sprintf("Model to use (deprecated, use --provider). Default: %s", generate.DefaultModel))
	generateCmd.Flags().Bool("list-models", false, "List all available models and exit")
	generateCmd.Flags().Bool("list-providers", false, "List all available providers with pricing and auth status")
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
