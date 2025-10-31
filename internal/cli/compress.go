package cli

import (
	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress [input]",
	Short: "Compress an image file",
	Long: `Compress an image file to reduce file size while maintaining quality.

Examples:
  gimage compress input.jpg --quality 85
  gimage compress input.png --output compressed.png`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement compress functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)

	// Flags for compress command
	compressCmd.Flags().StringP("output", "o", "", "output file path")
	compressCmd.Flags().Int("quality", 90, "compression quality (1-100)")
}
