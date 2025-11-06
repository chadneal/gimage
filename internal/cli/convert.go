package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apresai/gimage/internal/imaging"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert an image to a different format",
	Long: `Convert an image to a different format (PNG, JPG, WebP, GIF, TIFF, BMP).

Supported formats:
  • PNG  - Lossless, transparency support
  • JPG  - Lossy, best for photos
  • WebP - Modern format, smaller files
  • GIF  - Animated images, limited colors
  • TIFF - High quality, large files
  • BMP  - Uncompressed, largest files

Examples:
  gimage convert --input input.png --format jpg
  gimage convert -i input.jpg -f webp --output converted.webp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		inputPath, _ := cmd.Flags().GetString("input")
		targetFormat, _ := cmd.Flags().GetString("format")
		outputPath, _ := cmd.Flags().GetString("output")

		// Validate required flags
		if inputPath == "" {
			return fmt.Errorf("--input flag is required")
		}
		if targetFormat == "" {
			return fmt.Errorf("--format flag is required")
		}

		// Generate output path if not provided
		if outputPath == "" {
			ext := filepath.Ext(inputPath)
			base := inputPath[:len(inputPath)-len(ext)]
			outputPath = fmt.Sprintf("%s_converted.%s", base, targetFormat)
		}

		// Validate input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		}

		// Convert the image
		printInfo("Converting %s to %s format...", inputPath, targetFormat)
		printVerbose("Input: %s", inputPath)
		printVerbose("Output: %s", outputPath)
		printVerbose("Target format: %s", targetFormat)
		err := imaging.ConvertImageFile(context.Background(), inputPath, outputPath)
		if err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		// Get file size for reporting
		info, err := os.Stat(outputPath)
		if err != nil {
			return fmt.Errorf("failed to stat output file: %w", err)
		}

		printSuccess("Converted successfully!")
		printInfo("Output: %s", outputPath)
		printInfo("Size: %d bytes", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Flags for convert command
	convertCmd.Flags().StringP("input", "i", "", "input image file path (required)")
	convertCmd.Flags().StringP("format", "f", "", "target format: png, jpg, webp, gif, tiff, bmp (required)")
	convertCmd.Flags().StringP("output", "o", "", "output file path (default: input_converted.FORMAT)")

	convertCmd.MarkFlagRequired("input")
	convertCmd.MarkFlagRequired("format")
}
