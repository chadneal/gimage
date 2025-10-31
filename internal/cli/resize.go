package cli

import (
	"github.com/spf13/cobra"
)

// resizeCmd represents the resize command
var resizeCmd = &cobra.Command{
	Use:   "resize [input] [width] [height]",
	Short: "Resize an image to specific dimensions",
	Long: `Resize an image to specific dimensions using high-quality Lanczos resampling.

Examples:
  gimage resize input.jpg 800 600
  gimage resize input.png 1920 1080 --output resized.png`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement resize functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resizeCmd)

	// Flags for resize command
	resizeCmd.Flags().StringP("output", "o", "", "output file path")
}
