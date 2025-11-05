package cli

import (
	"context"

	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/apresai/gimage/internal/imaging"
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
		inputPath := args[0]

		// Parse x coordinate
		x, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid x coordinate '%s': must be an integer", args[1])
		}

		// Parse y coordinate
		y, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid y coordinate '%s': must be an integer", args[2])
		}

		// Parse width
		width, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid width '%s': must be a positive integer", args[3])
		}

		// Parse height
		height, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid height '%s': must be a positive integer", args[4])
		}

		// Get output path from flag or generate default
		outputPath, _ := cmd.Flags().GetString("output")
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
		fmt.Printf("Cropping %s to region (%d,%d) %dx%d...\n", inputPath, x, y, width, height)
		err = imaging.CropImage(context.Background(), inputPath, outputPath, x, y, width, height)
		if err != nil {
			return fmt.Errorf("crop failed: %w", err)
		}

		// Report success
		info, _ := os.Stat(outputPath)
		fmt.Printf("✓ Cropped successfully!\n")
		fmt.Printf("  Output: %s\n", outputPath)
		fmt.Printf("  Size: %d bytes\n", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cropCmd)

	// Flags for crop command
	cropCmd.Flags().StringP("output", "o", "", "output file path")
}
