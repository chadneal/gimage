package cli

import (
	"github.com/apresai/gimage/internal/tui"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive Terminal User Interface",
	Long: `Launch the gimage Terminal User Interface (TUI) for an interactive experience.

The TUI provides a menu-driven interface for:
  - Generating images from text prompts
  - Processing images (resize, scale, crop, compress, convert)
  - Configuring API keys and settings

Examples:
  gimage tui              # Launch the TUI
  gimage --interactive    # Alternative way to launch TUI`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
