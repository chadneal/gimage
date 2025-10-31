package cli

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gimage configuration",
	Long: `Manage gimage configuration including API keys and default settings.

Examples:
  gimage config set gemini-api-key YOUR_API_KEY
  gimage config get gemini-api-key
  gimage config list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement config functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
