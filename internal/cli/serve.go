package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chadneal/gimage/internal/config"
	"github.com/chadneal/gimage/internal/mcp"
	"github.com/chadneal/gimage/internal/mcp/tools"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start the gimage MCP (Model Context Protocol) server.

This allows AI assistants like Claude to use gimage for image generation
and processing operations. The server communicates over stdio using the
MCP protocol.

USAGE WITH CLAUDE DESKTOP:

Add this to your Claude Desktop MCP configuration:

macOS/Linux:
  {
    "mcpServers": {
      "gimage": {
        "command": "gimage",
        "args": ["serve"]
      }
    }
  }

Configuration file location:
  • macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
  • Linux: ~/.config/Claude/claude_desktop_config.json
  • Windows: %APPDATA%\Claude\claude_desktop_config.json

ENVIRONMENT VARIABLES:

The serve command respects the same environment variables as the CLI:
  • GEMINI_API_KEY - Gemini API key for image generation
  • VERTEX_API_KEY - Vertex AI API key (Express Mode)
  • VERTEX_PROJECT - GCP project ID for Vertex AI
  • VERTEX_LOCATION - Vertex AI location (default: us-central1)
  • GOOGLE_APPLICATION_CREDENTIALS - Path to service account JSON

FEATURES:

The MCP server exposes 10 tools to AI assistants:
  • generate_image    - AI image generation with Gemini/Vertex
  • resize_image      - Resize to specific dimensions
  • scale_image       - Scale by factor
  • crop_image        - Crop to region
  • compress_image    - Compress to reduce file size
  • convert_image     - Convert between formats
  • batch_resize      - Batch resize operations
  • batch_compress    - Batch compression
  • batch_convert     - Batch format conversion
  • list_models       - List available AI models

EXAMPLES:

  # Start MCP server (usually called by AI assistant)
  $ gimage serve

  # Test with verbose logging (logs go to stderr)
  $ gimage serve --verbose

  # Use custom config file
  $ gimage serve --config ~/.gimage/custom-config.yaml

TROUBLESHOOTING:

If the MCP server isn't working in Claude:
  1. Check that gimage is in your PATH: which gimage
  2. Verify your API keys are configured: gimage auth gemini
  3. Look at Claude's logs for error messages
  4. Test image generation works: gimage generate "test image"

For more information: https://github.com/chadneal/gimage`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration (optional - tools will load it when needed)
		cfg, err := config.LoadConfig()
		if err != nil {
			// Non-fatal error - server can still work without config for some operations
			fmt.Fprintf(os.Stderr, "[gimage-mcp] Warning: failed to load config: %v\n", err)
			cfg = &config.Config{} // Use empty config
		}

		// Get verbose flag
		verbose := viper.GetBool("verbose")

		// Create MCP server
		server := mcp.NewMCPServer("gimage", version, cfg, verbose)

		// Register all tools
		tools.RegisterGenerateImageTool(server)
		tools.RegisterResizeImageTool(server)
		tools.RegisterScaleImageTool(server)
		tools.RegisterCropImageTool(server)
		tools.RegisterCompressImageTool(server)
		tools.RegisterConvertImageTool(server)
		tools.RegisterBatchResizeTool(server)
		tools.RegisterBatchCompressTool(server)
		tools.RegisterBatchConvertTool(server)
		tools.RegisterListModelsTool(server)

		// Setup signal handling for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Fprintln(os.Stderr, "\n[gimage-mcp] Shutting down gracefully...")
			cancel()
		}()

		// Log startup (to stderr, not stdout - stdout is for JSON-RPC)
		if verbose {
			fmt.Fprintln(os.Stderr, "[gimage-mcp] Starting MCP server")
			fmt.Fprintln(os.Stderr, "[gimage-mcp] Protocol: Model Context Protocol")
			fmt.Fprintln(os.Stderr, "[gimage-mcp] Transport: stdio")
			fmt.Fprintln(os.Stderr, "[gimage-mcp] Tools: 10 registered")
			fmt.Fprintln(os.Stderr, "[gimage-mcp] Ready for requests")
		}

		// Start server
		if err := server.Start(ctx); err != nil && err != context.Canceled {
			return fmt.Errorf("server error: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
