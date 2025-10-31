package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chadneal/gimage/internal/config"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication for Gemini and Vertex AI",
	Long: `Interactive authentication setup for Gemini API and Vertex AI.

This command helps you configure your API credentials securely.`,
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

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authGeminiCmd)
	authCmd.AddCommand(authVertexCmd)
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

// maskString masks all but the last 4 characters of a string
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "..." + s[len(s)-4:]
}
