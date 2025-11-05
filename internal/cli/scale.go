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
		inputPath := args[0]

		// Parse factor argument
		factor, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid factor '%s': must be a positive number", args[1])
		}

		// Get output path from flag or generate default
		outputPath, _ := cmd.Flags().GetString("output")
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
		fmt.Printf("Scaling %s by %.2fx...\n", inputPath, factor)
		err = imaging.ScaleImage(context.Background(), inputPath, outputPath, factor)
		if err != nil {
			return fmt.Errorf("scale failed: %w", err)
		}

		// Report success
		info, _ := os.Stat(outputPath)
		fmt.Printf("✓ Scaled successfully!\n")
		fmt.Printf("  Output: %s\n", outputPath)
		fmt.Printf("  Size: %d bytes\n", info.Size())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scaleCmd)

	// Flags for scale command
	scaleCmd.Flags().StringP("output", "o", "", "output file path")
}
