package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apresai/gimage/internal/imaging"
	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Compress an image file to reduce size",
	Long: `Compress an image file to reduce file size while maintaining quality.

SUPPORTED FORMATS (with quality control):
  • JPEG/JPG  - Lossy compression with quality 1-100
  • WebP      - Lossy compression with quality 1-100

UNSUPPORTED FORMATS (copy only):
  • PNG       - Already compressed losslessly (copy only)
  • GIF       - Already compressed (copy only)
  • TIFF      - Typically uncompressed (copy only)
  • BMP       - Uncompressed format (copy only)

Quality Guide:
  • 90-100: Excellent quality, larger file size
  • 85-90:  Very good quality, balanced (default: 85)
  • 75-85:  Good quality, smaller file size
  • 60-75:  Acceptable quality, much smaller
  • Below 60: Noticeable quality loss

Examples:
  # Compress JPEG with default quality (85)
  gimage compress --input photo.jpg --output compressed.jpg

  # Compress with high quality
  gimage compress --input image.jpg --quality 90 --output high-quality.jpg

  # Compress WebP with lower quality for web
  gimage compress --input photo.webp --quality 75 --output web-optimized.webp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		inputPath, _ := cmd.Flags().GetString("input")
		outputPath, _ := cmd.Flags().GetString("output")
		quality, _ := cmd.Flags().GetInt("quality")

		// Validate input
		if inputPath == "" {
			return fmt.Errorf("--input flag is required")
		}

		// Validate input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		}

		// Generate output path if not provided
		if outputPath == "" {
			ext := filepath.Ext(inputPath)
			base := inputPath[:len(inputPath)-len(ext)]
			outputPath = fmt.Sprintf("%s_compressed%s", base, ext)
		}

		// Get file extension to check format
		ext := filepath.Ext(inputPath)
		format := ext[1:] // Remove leading dot

		// Check if format supports compression
		if format != "jpg" && format != "jpeg" && format != "webp" {
			printWarning("Format '%s' does not support quality-based compression", format)
			printInfo("Supported formats: JPG, JPEG, WebP")
			printInfo("The file will be copied without compression")
			printInfo("")

			// Just copy the file
			return copyFile(inputPath, outputPath)
		}

		// Perform compression
		printInfo("Compressing image...")
		printVerbose("Input: %s", inputPath)
		printVerbose("Output: %s", outputPath)
		printVerbose("Format: %s", format)
		printVerbose("Quality: %d", quality)

		ctx := context.Background()
		err := imaging.CompressImage(ctx, inputPath, outputPath, quality)
		if err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}

		// Get file sizes for comparison
		originalSize, _ := getFileSize(inputPath)
		compressedSize, _ := getFileSize(outputPath)

		if originalSize > 0 && compressedSize > 0 {
			savings := originalSize - compressedSize
			savingsPercent := (float64(savings) / float64(originalSize)) * 100

			printSuccess("Image compressed successfully!")
			printInfo("Original size:   %s", formatFileSize(originalSize))
			printInfo("Compressed size: %s", formatFileSize(compressedSize))
			if savings > 0 {
				printInfo("Saved:           %s (%.1f%%)", formatFileSize(savings), savingsPercent)
			} else {
				printWarning("Note: Compressed file is larger (quality may be too high)")
			}
			printInfo("Output: %s", outputPath)
		} else {
			printSuccess("Image compressed successfully!")
			printInfo("Output: %s", outputPath)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)

	// Flags for compress command
	compressCmd.Flags().StringP("input", "i", "", "input image file path (required)")
	compressCmd.Flags().StringP("output", "o", "", "output file path (default: input_compressed.ext)")
	compressCmd.Flags().IntP("quality", "q", 85, "compression quality 1-100 (default: 85)")
	compressCmd.MarkFlagRequired("input")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	printSuccess("File copied successfully (no compression applied)")
	printInfo("Output: %s", dst)
	return nil
}

// getFileSize returns the size of a file in bytes
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// formatFileSize formats bytes as human-readable string
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
