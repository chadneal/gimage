package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apresai/gimage/internal/imaging"
	"github.com/spf13/cobra"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "Scale an image by a factor",
	Long: `Scale an image by a factor (e.g., 0.5 for half size, 2.0 for double size).

Examples:
  gimage scale --input input.jpg --factor 0.5
  gimage scale -i input.png -f 2.0 --output scaled.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		inputPath, _ := cmd.Flags().GetString("input")
		factor, _ := cmd.Flags().GetFloat64("factor")
		outputPath, _ := cmd.Flags().GetString("output")

		// Validate required flags
		if inputPath == "" {
			return fmt.Errorf("--input flag is required")
		}
		if factor <= 0 {
			return fmt.Errorf("--factor must be a positive number")
		}

		// Generate output path if not provided
		if outputPath == "" {
			ext := filepath.Ext(inputPath)
			base := inputPath[:len(inputPath)-len(ext)]
			outputPath = fmt.Sprintf("%s_scaled_%.2fx%s", base, factor, ext)
		}

		// Validate input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		}

		// Call imaging function
		printInfo("Scaling %s by %.2fx...", inputPath, factor)
		printVerbose("Input: %s", inputPath)
		printVerbose("Output: %s", outputPath)
		printVerbose("Scale factor: %.2f", factor)
		err := imaging.ScaleImage(context.Background(), inputPath, outputPath, factor)
		if err != nil {
			return fmt.Errorf("scale failed: %w", err)
		}

		// Report success
		info, _ := os.Stat(outputPath)
		printSuccess("Scaled successfully!")
		printInfo("Output: %s", outputPath)
		printInfo("Size: %d bytes", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scaleCmd)

	// Flags for scale command
	scaleCmd.Flags().StringP("input", "i", "", "input image file path (required)")
	scaleCmd.Flags().Float64P("factor", "f", 0, "scale factor (e.g., 0.5 = half, 2.0 = double) (required)")
	scaleCmd.Flags().StringP("output", "o", "", "output file path (default: input_scaled_FACTORx.ext)")

	scaleCmd.MarkFlagRequired("input")
	scaleCmd.MarkFlagRequired("factor")
}
