package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/apresai/gimage/internal/config"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication for all image generation providers",
	Long: `Manage authentication credentials for all supported providers.

Each provider (Gemini API, Vertex AI, AWS Bedrock) has specific credential
requirements and pricing. Use the subcommands to:

  list  - View all providers and their authentication status
  test  - Test if authentication works for a provider
  setup - Interactively configure credentials for a provider

Different providers offer different models with different pricing:
  - gemini/flash-2.5: FREE tier via Gemini API (500/day)
  - vertex/imagen-4: $0.04/image via Vertex AI
  - bedrock/nova-canvas: $0.08/image via AWS Bedrock`,
	Example: `  # List all providers and auth status
  gimage auth list

  # Set up Gemini authentication
  gimage auth setup gemini

  # Test if Vertex AI works
  gimage auth test vertex/imagen-4`,
}

// authGeminiCmd handles Gemini API authentication
var authGeminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Configure Gemini API authentication",
	Long: `Interactive setup for Gemini API key.

Get your API key from: https://aistudio.google.com/app/apikey`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupGeminiAuth()
	},
}

// authVertexCmd handles Vertex AI authentication
var authVertexCmd = &cobra.Command{
	Use:   "vertex",
	Short: "Configure Vertex AI authentication",
	Long: `Interactive setup for Vertex AI credentials.

Vertex AI supports two authentication modes:
  1. Express Mode - Simple API key (good for testing/development)
  2. Full Mode - Service account (for production)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupVertexAuth()
	},
}

// authBedrockCmd handles AWS Bedrock authentication
var authBedrockCmd = &cobra.Command{
	Use:   "bedrock",
	Short: "Configure AWS Bedrock authentication",
	Long: `Interactive setup for AWS Bedrock credentials.

AWS Bedrock supports multiple authentication methods:
  1. Access Keys - AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
  2. AWS Profile - Named profile from ~/.aws/credentials
  3. IAM Role - Automatic for EC2/ECS/Lambda (no setup needed)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupBedrockAuth()
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	// New provider-based commands
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authTestCmd)
	authCmd.AddCommand(authSetupCmd)
	// Legacy commands (kept for backward compatibility)
	authCmd.AddCommand(authGeminiCmd)
	authCmd.AddCommand(authVertexCmd)
	authCmd.AddCommand(authBedrockCmd)
}

// setupGeminiAuth runs the interactive setup for Gemini API
func setupGeminiAuth() error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config to use as defaults
	existingCfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist yet, create a new one with defaults
		existingCfg = &config.Config{
			DefaultAPI:     "gemini",
			DefaultModel:   "gemini-2.5-flash-image",
			DefaultSize:    "1024x1024",
			VertexLocation: "us-central1",
			LogLevel:       "info",
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Gemini API Authentication Setup")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Get your API key from: https://aistudio.google.com/app/apikey")
	fmt.Println()

	// Gemini API Key
	apiKey := promptWithDefault(reader, "Gemini API Key", existingCfg.GeminiAPIKey, true)

	// Update config
	existingCfg.GeminiAPIKey = apiKey

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ Configuration saved successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use Gemini API with:")
	fmt.Println("  gimage generate \"your prompt here\"")
	fmt.Println()

	return nil
}

// setupVertexAuth runs the interactive setup for Vertex AI
func setupVertexAuth() error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config to use as defaults
	existingCfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist yet, create a new one with defaults
		existingCfg = &config.Config{
			DefaultAPI:     "gemini",
			DefaultModel:   "gemini-2.5-flash-image",
			DefaultSize:    "1024x1024",
			VertexLocation: "us-central1",
			LogLevel:       "info",
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Vertex AI Authentication Setup")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Choose your authentication mode:")
	fmt.Println()
	fmt.Println("  1. Express Mode - API Key (simple, good for testing)")
	fmt.Println("     • Sign up at: https://console.cloud.google.com/vertex-ai")
	fmt.Println("     • Get API key from: APIs & Services > Credentials")
	fmt.Println("     • Best for: Development, testing, rapid prototyping")
	fmt.Println()
	fmt.Println("  2. Full Mode - Service Account (secure, production-ready)")
	fmt.Println("     • Requires: GCP project, service account JSON file")
	fmt.Println("     • Best for: Production, fine-grained access control")
	fmt.Println()
	fmt.Println("  3. Full Mode - Application Default Credentials (local dev)")
	fmt.Println("     • Run: gcloud auth application-default login")
	fmt.Println("     • Best for: Local development with your GCP account")
	fmt.Println()

	// Default to express mode if they already have an API key
	defaultMode := "1"
	if existingCfg.VertexCredentialsPath != "" {
		defaultMode = "2"
	} else if existingCfg.VertexAPIKey != "" {
		defaultMode = "1"
	}

	mode := promptWithDefault(reader, "Choose mode (1, 2, or 3)", defaultMode, false)

	fmt.Println()

	if mode == "1" {
		// Express Mode - API Key
		return setupVertexExpressMode(reader, existingCfg)
	} else if mode == "2" {
		// Full Mode - Service Account
		return setupVertexFullModeServiceAccount(reader, existingCfg)
	} else {
		// Full Mode - ADC
		return setupVertexFullModeADC(reader, existingCfg)
	}
}

// setupVertexExpressMode sets up Vertex AI with API key (express mode)
func setupVertexExpressMode(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ Express Mode Setup ━━━")
	fmt.Println()
	fmt.Println("Get your API key:")
	fmt.Println("  1. Go to: https://console.cloud.google.com/vertex-ai")
	fmt.Println("  2. Sign up for Vertex AI Express Mode")
	fmt.Println("  3. Find your API key in: APIs & Services > Credentials")
	fmt.Println()

	// API Key
	apiKey := promptWithDefault(reader, "Vertex AI API Key", existingCfg.VertexAPIKey, true)

	// Project ID (optional for express mode, but good to have)
	projectID := promptWithDefault(reader, "Google Cloud Project ID (optional)", existingCfg.VertexProject, false)

	// Location/Region
	location := promptWithDefault(reader, "Location/Region", existingCfg.VertexLocation, false)

	// Update config
	existingCfg.VertexAPIKey = apiKey
	existingCfg.VertexProject = projectID
	existingCfg.VertexLocation = location
	existingCfg.VertexCredentialsPath = "" // Clear service account path

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ Express Mode configured successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use Vertex AI with:")
	fmt.Println("  gimage generate --api vertex \"your prompt here\"")
	fmt.Println()

	return nil
}

// setupVertexFullModeServiceAccount sets up Vertex AI with service account
func setupVertexFullModeServiceAccount(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ Full Mode - Service Account Setup ━━━")
	fmt.Println()

	// Project ID
	projectID := promptWithDefault(reader, "Google Cloud Project ID", existingCfg.VertexProject, false)

	// Location/Region
	location := promptWithDefault(reader, "Location/Region", existingCfg.VertexLocation, false)

	// Service account JSON file path
	defaultPath := existingCfg.VertexCredentialsPath
	credsPath := promptWithDefault(reader, "Path to service account JSON file", defaultPath, false)

	// Expand home directory
	if strings.HasPrefix(credsPath, "~/") {
		home, _ := os.UserHomeDir()
		credsPath = strings.Replace(credsPath, "~", home, 1)
	}

	// Verify file exists
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		return fmt.Errorf("service account file not found: %s", credsPath)
	}

	// Update config
	existingCfg.VertexProject = projectID
	existingCfg.VertexLocation = location
	existingCfg.VertexCredentialsPath = credsPath
	existingCfg.VertexAPIKey = "" // Clear API key

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ Full Mode configured successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use Vertex AI with:")
	fmt.Println("  gimage generate --api vertex \"your prompt here\"")
	fmt.Println()

	return nil
}

// setupVertexFullModeADC sets up Vertex AI with Application Default Credentials
func setupVertexFullModeADC(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ Full Mode - Application Default Credentials Setup ━━━")
	fmt.Println()

	// Project ID
	projectID := promptWithDefault(reader, "Google Cloud Project ID", existingCfg.VertexProject, false)

	// Location/Region
	location := promptWithDefault(reader, "Location/Region", existingCfg.VertexLocation, false)

	fmt.Println()
	fmt.Println("Using Application Default Credentials (ADC).")
	fmt.Println()
	fmt.Println("Make sure you've run:")
	fmt.Println("  gcloud auth application-default login")
	fmt.Println()

	// Update config
	existingCfg.VertexProject = projectID
	existingCfg.VertexLocation = location
	existingCfg.VertexCredentialsPath = "" // Empty means use ADC
	existingCfg.VertexAPIKey = ""          // Clear API key

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ Full Mode (ADC) configured successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use Vertex AI with:")
	fmt.Println("  gimage generate --api vertex \"your prompt here\"")
	fmt.Println()

	return nil
}

// promptWithDefault prompts the user for input with a default value
// If hideInput is true, the input will be hidden (for API keys)
func promptWithDefault(reader *bufio.Reader, prompt, defaultValue string, hideInput bool) string {
	if defaultValue != "" {
		if hideInput {
			// Show masked version of default
			masked := maskString(defaultValue)
			fmt.Printf("%s [%s]: ", prompt, masked)
		} else {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		}
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	return input
}

// setupBedrockAuth runs the interactive setup for AWS Bedrock
func setupBedrockAuth() error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config to use as defaults
	existingCfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist yet, create a new one with defaults
		existingCfg = &config.Config{
			DefaultAPI:     "gemini",
			DefaultModel:   "gemini-2.5-flash-image",
			DefaultSize:    "1024x1024",
			VertexLocation: "us-central1",
			AWSRegion:      "us-east-1",
			LogLevel:       "info",
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  AWS Bedrock Authentication Setup")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Choose your authentication method:")
	fmt.Println()
	fmt.Println("  1. Bedrock API Key (Bearer Token) - Simple REST API authentication")
	fmt.Println("     • Get from AWS Console > Bedrock > API Keys")
	fmt.Println("     • Best for: Quick setup, development")
	fmt.Println()
	fmt.Println("  2. Access Keys - Direct AWS credentials (SDK)")
	fmt.Println("     • Enter AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
	fmt.Println("     • Best for: Testing, development")
	fmt.Println()
	fmt.Println("  3. AWS Profile - Named profile from ~/.aws/credentials (SDK)")
	fmt.Println("     • Use an existing AWS CLI profile")
	fmt.Println("     • Best for: Multiple AWS accounts, organized credentials")
	fmt.Println()
	fmt.Println("  4. IAM Role - Use instance/container role (SDK, no setup needed)")
	fmt.Println("     • Automatic for EC2, ECS, Lambda environments")
	fmt.Println("     • Best for: Production deployments, security")
	fmt.Println()

	// Default to bearer token if it exists
	defaultMode := "1"
	if existingCfg.AWSBedrockAPIKey != "" {
		defaultMode = "1"
	} else if existingCfg.AWSAccessKeyID != "" {
		defaultMode = "2"
	} else if existingCfg.AWSProfile != "" {
		defaultMode = "3"
	}

	mode := promptWithDefault(reader, "Choose mode (1, 2, 3, or 4)", defaultMode, false)

	fmt.Println()

	if mode == "1" {
		// Bedrock API Key (Bearer Token)
		return setupBedrockAPIKey(reader, existingCfg)
	} else if mode == "2" {
		// Access Keys
		return setupBedrockAccessKeys(reader, existingCfg)
	} else if mode == "3" {
		// AWS Profile
		return setupBedrockProfile(reader, existingCfg)
	} else {
		// IAM Role - no config needed
		return setupBedrockIAMRole(reader, existingCfg)
	}
}

// setupBedrockAPIKey sets up AWS Bedrock with API key (bearer token)
func setupBedrockAPIKey(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ Bedrock API Key Setup ━━━")
	fmt.Println()
	fmt.Println("Get your AWS Bedrock API key from:")
	fmt.Println("  AWS Console > Bedrock > API Keys")
	fmt.Println()
	fmt.Println("This uses REST API authentication (simpler, recommended for development)")
	fmt.Println()

	// API Key
	apiKey := promptWithDefault(reader, "AWS Bedrock API Key", existingCfg.AWSBedrockAPIKey, true)

	// AWS Region
	region := promptWithDefault(reader, "AWS Region", existingCfg.AWSRegion, false)

	// Update config
	existingCfg.AWSBedrockAPIKey = apiKey
	existingCfg.AWSRegion = region
	existingCfg.AWSAccessKeyID = ""     // Clear SDK credentials
	existingCfg.AWSSecretAccessKey = "" // Clear SDK credentials
	existingCfg.AWSProfile = ""         // Clear profile

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ AWS Bedrock configured successfully (REST API mode)!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use AWS Bedrock with:")
	fmt.Println("  gimage generate --api bedrock \"your prompt here\"")
	fmt.Println()
	fmt.Println("Or use the nova-canvas model alias:")
	fmt.Println("  gimage generate --model nova-canvas \"your prompt here\"")
	fmt.Println()

	return nil
}

// setupBedrockAccessKeys sets up AWS Bedrock with access keys
func setupBedrockAccessKeys(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ Access Keys Setup ━━━")
	fmt.Println()
	fmt.Println("Get your AWS credentials from:")
	fmt.Println("  AWS Console > IAM > Users > Security Credentials")
	fmt.Println()
	fmt.Println("Required IAM permissions:")
	fmt.Println("  • bedrock:InvokeModel")
	fmt.Println()

	// Access Key ID
	accessKeyID := promptWithDefault(reader, "AWS Access Key ID", existingCfg.AWSAccessKeyID, false)

	// Secret Access Key
	secretKey := promptWithDefault(reader, "AWS Secret Access Key", existingCfg.AWSSecretAccessKey, true)

	// AWS Region
	region := promptWithDefault(reader, "AWS Region", existingCfg.AWSRegion, false)

	// Update config
	existingCfg.AWSAccessKeyID = accessKeyID
	existingCfg.AWSSecretAccessKey = secretKey
	existingCfg.AWSRegion = region
	existingCfg.AWSProfile = ""        // Clear profile if using access keys
	existingCfg.AWSBedrockAPIKey = "" // Clear bearer token if using SDK

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ AWS Bedrock configured successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use AWS Bedrock with:")
	fmt.Println("  gimage generate --api bedrock \"your prompt here\"")
	fmt.Println()
	fmt.Println("Or use the nova-canvas model alias:")
	fmt.Println("  gimage generate --model nova-canvas \"your prompt here\"")
	fmt.Println()

	return nil
}

// setupBedrockProfile sets up AWS Bedrock with AWS profile
func setupBedrockProfile(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ AWS Profile Setup ━━━")
	fmt.Println()
	fmt.Println("Using AWS CLI profile from ~/.aws/credentials")
	fmt.Println()

	// AWS Profile name
	profile := promptWithDefault(reader, "AWS Profile name", existingCfg.AWSProfile, false)

	// AWS Region
	region := promptWithDefault(reader, "AWS Region", existingCfg.AWSRegion, false)

	// Update config
	existingCfg.AWSProfile = profile
	existingCfg.AWSRegion = region
	existingCfg.AWSAccessKeyID = ""     // Clear access keys if using profile
	existingCfg.AWSSecretAccessKey = "" // Clear access keys if using profile
	existingCfg.AWSBedrockAPIKey = ""   // Clear bearer token if using SDK

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ AWS Bedrock configured successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("You can now use AWS Bedrock with:")
	fmt.Println("  gimage generate --api bedrock \"your prompt here\"")
	fmt.Println()

	return nil
}

// setupBedrockIAMRole sets up AWS Bedrock with IAM role (no credentials needed)
func setupBedrockIAMRole(reader *bufio.Reader, existingCfg *config.Config) error {
	fmt.Println("━━━ IAM Role Setup ━━━")
	fmt.Println()
	fmt.Println("Using IAM role from EC2/ECS/Lambda instance.")
	fmt.Println()
	fmt.Println("No credentials needed - the AWS SDK will automatically")
	fmt.Println("use the instance's IAM role.")
	fmt.Println()

	// AWS Region (still needed)
	region := promptWithDefault(reader, "AWS Region", existingCfg.AWSRegion, false)

	// Update config
	existingCfg.AWSRegion = region
	existingCfg.AWSAccessKeyID = ""     // Clear access keys
	existingCfg.AWSSecretAccessKey = "" // Clear secret key
	existingCfg.AWSProfile = ""         // Clear profile
	existingCfg.AWSBedrockAPIKey = ""   // Clear bearer token if using SDK

	// Save config
	if err := config.SaveConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath := config.GetConfigPath()
	fmt.Println()
	fmt.Println("✓ AWS Bedrock configured successfully!")
	fmt.Printf("  Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("Make sure your instance has an IAM role with:")
	fmt.Println("  • bedrock:InvokeModel permission")
	fmt.Println("  • Model access enabled in AWS Bedrock console")
	fmt.Println()
	fmt.Println("You can now use AWS Bedrock with:")
	fmt.Println("  gimage generate --api bedrock \"your prompt here\"")
	fmt.Println()

	return nil
}

// maskString masks all but the last 4 characters of a string
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "..." + s[len(s)-4:]
}
