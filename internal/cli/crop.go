package cli

import (
	"github.com/spf13/cobra"
)

// cropCmd represents the crop command
var cropCmd = &cobra.Command{
	Use:   "crop [input] [x] [y] [width] [height]",
	Short: "Crop an image to a specific region",
	Long: `Crop an image to a specific region defined by x, y coordinates and dimensions.

Examples:
  gimage crop input.jpg 100 100 800 600
  gimage crop input.png 0 0 1920 1080 --output cropped.png`,
	Args: cobra.ExactArgs(5),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement crop functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cropCmd)

	// Flags for crop command
	cropCmd.Flags().StringP("output", "o", "", "output file path")
}
