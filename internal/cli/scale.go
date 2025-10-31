package cli

import (
	"github.com/spf13/cobra"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale [input] [factor]",
	Short: "Scale an image by a factor",
	Long: `Scale an image by a factor (e.g., 0.5 for half size, 2.0 for double size).

Examples:
  gimage scale input.jpg 0.5
  gimage scale input.png 2.0 --output scaled.png`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement scale functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scaleCmd)

	// Flags for scale command
	scaleCmd.Flags().StringP("output", "o", "", "output file path")
}
