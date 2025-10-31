package cli

import (
	"github.com/spf13/cobra"
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch [operation] [input-dir]",
	Short: "Batch process multiple images",
	Long: `Batch process multiple images with concurrent operations.

Examples:
  gimage batch resize images/ --width 800 --height 600
  gimage batch compress images/ --quality 85 --output processed/
  gimage batch convert images/ jpg`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement batch functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)

	// Flags for batch command
	batchCmd.Flags().StringP("output", "o", "", "output directory path")
	batchCmd.Flags().Int("workers", 4, "number of parallel workers")
	batchCmd.Flags().Int("width", 0, "width for resize operation")
	batchCmd.Flags().Int("height", 0, "height for resize operation")
	batchCmd.Flags().Int("quality", 90, "quality for compress operation")
}
