package cli

import (
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert [input] [format]",
	Short: "Convert an image to a different format",
	Long: `Convert an image to a different format (PNG, JPG, WebP, GIF, TIFF, BMP).

Examples:
  gimage convert input.png jpg
  gimage convert input.jpg webp --output converted.webp`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement convert functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Flags for convert command
	convertCmd.Flags().StringP("output", "o", "", "output file path")
}
