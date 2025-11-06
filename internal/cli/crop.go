package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apresai/gimage/internal/imaging"
	"github.com/spf13/cobra"
)

// cropCmd represents the crop command
var cropCmd = &cobra.Command{
	Use:   "crop",
	Short: "Crop an image to a specific region",
	Long: `Crop an image to a specific region defined by x, y coordinates and dimensions.

Examples:
  gimage crop --input input.jpg --x 100 --y 100 --width 800 --height 600
  gimage crop -i input.png --x 0 --y 0 -w 1920 -h 1080 --output cropped.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		inputPath, _ := cmd.Flags().GetString("input")
		x, _ := cmd.Flags().GetInt("x")
		y, _ := cmd.Flags().GetInt("y")
		width, _ := cmd.Flags().GetInt("width")
		height, _ := cmd.Flags().GetInt("height")
		outputPath, _ := cmd.Flags().GetString("output")

		// Validate required flags
		if inputPath == "" {
			return fmt.Errorf("--input flag is required")
		}
		if width <= 0 {
			return fmt.Errorf("--width must be a positive integer")
		}
		if height <= 0 {
			return fmt.Errorf("--height must be a positive integer")
		}

		// Generate output path if not provided
		if outputPath == "" {
			ext := filepath.Ext(inputPath)
			base := inputPath[:len(inputPath)-len(ext)]
			outputPath = fmt.Sprintf("%s_cropped_%dx%d%s", base, width, height, ext)
		}

		// Validate input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		}

		// Call imaging function
		printInfo("Cropping %s to region (%d,%d) %dx%d...", inputPath, x, y, width, height)
		printVerbose("Input: %s", inputPath)
		printVerbose("Output: %s", outputPath)
		printVerbose("Region: x=%d, y=%d, width=%d, height=%d", x, y, width, height)
		err := imaging.CropImage(context.Background(), inputPath, outputPath, x, y, width, height)
		if err != nil {
			return fmt.Errorf("crop failed: %w", err)
		}

		// Report success
		info, _ := os.Stat(outputPath)
		printSuccess("Cropped successfully!")
		printInfo("Output: %s", outputPath)
		printInfo("Size: %d bytes", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cropCmd)

	// Flags for crop command
	cropCmd.Flags().StringP("input", "i", "", "input image file path (required)")
	cropCmd.Flags().Int("x", 0, "x coordinate of top-left corner (default: 0)")
	cropCmd.Flags().Int("y", 0, "y coordinate of top-left corner (default: 0)")
	cropCmd.Flags().Int("width", 0, "width of crop region in pixels (required)")
	cropCmd.Flags().Int("height", 0, "height of crop region in pixels (required)")
	cropCmd.Flags().StringP("output", "o", "", "output file path (default: input_cropped_WxH.ext)")

	cropCmd.MarkFlagRequired("input")
	cropCmd.MarkFlagRequired("width")
	cropCmd.MarkFlagRequired("height")
}
