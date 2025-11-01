package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apresai/gimage/internal/imaging"
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
		inputPath := args[0]
		targetFormat := args[1]

		// Get output path from flag or generate default
		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			// Generate output path: input_converted.format
			ext := filepath.Ext(inputPath)
			base := inputPath[:len(inputPath)-len(ext)]
			outputPath = fmt.Sprintf("%s_converted.%s", base, targetFormat)
		}

		// Validate input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		}

		// Convert the image
		fmt.Printf("Converting %s to %s format...\n", inputPath, targetFormat)
		err := imaging.ConvertImageFile(inputPath, outputPath)
		if err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		// Get file size for reporting
		info, err := os.Stat(outputPath)
		if err != nil {
			return fmt.Errorf("failed to stat output file: %w", err)
		}

		fmt.Printf("✓ Converted successfully!\n")
		fmt.Printf("  Output: %s\n", outputPath)
		fmt.Printf("  Size: %d bytes\n", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Flags for convert command
	convertCmd.Flags().StringP("output", "o", "", "output file path")
}
