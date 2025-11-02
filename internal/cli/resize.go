package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/apresai/gimage/internal/imaging"
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
		inputPath := args[0]

		// Parse width and height
		width, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid width '%s': must be a positive integer", args[1])
		}
		height, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid height '%s': must be a positive integer", args[2])
		}

		// Get output path from flag or generate default
		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			// Generate output path: input_resized_WxH.ext
			ext := filepath.Ext(inputPath)
			base := inputPath[:len(inputPath)-len(ext)]
			outputPath = fmt.Sprintf("%s_resized_%dx%d%s", base, width, height, ext)
		}

		// Validate input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		}

		// Resize the image
		fmt.Printf("Resizing %s to %dx%d...\n", inputPath, width, height)
		err = imaging.ResizeImage(inputPath, outputPath, width, height)
		if err != nil {
			return fmt.Errorf("resize failed: %w", err)
		}

		// Get file size for reporting
		info, err := os.Stat(outputPath)
		if err != nil {
			return fmt.Errorf("failed to stat output file: %w", err)
		}

		fmt.Printf("✓ Resized successfully!\n")
		fmt.Printf("  Output: %s\n", outputPath)
		fmt.Printf("  Size: %d bytes\n", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(resizeCmd)

	// Flags for resize command
	resizeCmd.Flags().StringP("output", "o", "", "output file path")
}
