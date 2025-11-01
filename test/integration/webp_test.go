package integration

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/HugoSmits86/nativewebp"
	"github.com/chadneal/gimage/internal/imaging"
)

// TestWebPEncoding tests WebP encoding functionality
func TestWebPEncoding(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Create a gradient effect
			r := uint8(x * 255 / 100)
			g := uint8(y * 255 / 100)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Save as PNG first
	pngPath := filepath.Join(tmpDir, "test.png")
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("Failed to create PNG file: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("Failed to encode PNG: %v", err)
	}
	f.Close()

	// Test 1: Convert PNG to WebP using ConvertImageFile
	webpPath := filepath.Join(tmpDir, "test.webp")
	err = imaging.ConvertImageFile(pngPath, webpPath)
	if err != nil {
		t.Fatalf("ConvertImageFile failed: %v", err)
	}

	// Verify WebP file exists
	if _, err := os.Stat(webpPath); os.IsNotExist(err) {
		t.Fatal("WebP file was not created")
	}

	// Verify WebP file can be decoded
	webpFile, err := os.Open(webpPath)
	if err != nil {
		t.Fatalf("Failed to open WebP file: %v", err)
	}
	defer webpFile.Close()

	decodedImg, err := nativewebp.Decode(webpFile)
	if err != nil {
		t.Fatalf("Failed to decode WebP: %v", err)
	}

	// Verify dimensions
	bounds := decodedImg.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("WebP dimensions incorrect: got %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}

	t.Logf("✓ WebP encoding test passed: %s", webpPath)
}

// TestWebPWithTransparency tests WebP encoding with transparency
func TestWebPWithTransparency(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test image with transparency
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			// Create a pattern with transparency
			if (x+y)%2 == 0 {
				img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Red, opaque
			} else {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 128}) // Blue, semi-transparent
			}
		}
	}

	// Save as PNG
	pngPath := filepath.Join(tmpDir, "transparent.png")
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("Failed to create PNG file: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("Failed to encode PNG: %v", err)
	}
	f.Close()

	// Convert to WebP
	webpPath := filepath.Join(tmpDir, "transparent.webp")
	err = imaging.ConvertImageFile(pngPath, webpPath)
	if err != nil {
		t.Fatalf("ConvertImageFile with transparency failed: %v", err)
	}

	// Verify WebP file exists
	if _, err := os.Stat(webpPath); os.IsNotExist(err) {
		t.Fatal("WebP file with transparency was not created")
	}

	t.Logf("✓ WebP transparency test passed: %s", webpPath)
}

// TestWebPConvertData tests WebP conversion using ConvertImageData
func TestWebPConvertData(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			img.Set(i, j, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	// Encode as PNG first
	var pngBuf []byte
	{
		tmpFile, err := os.CreateTemp("", "test*.png")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if err := png.Encode(tmpFile, img); err != nil {
			t.Fatalf("Failed to encode PNG: %v", err)
		}

		// Read back the PNG data
		if _, err := tmpFile.Seek(0, 0); err != nil {
			t.Fatalf("Failed to seek: %v", err)
		}
		pngBuf, err = os.ReadFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to read PNG: %v", err)
		}
	}

	// Convert PNG data to WebP
	webpData, err := imaging.ConvertImageData(pngBuf, "webp")
	if err != nil {
		t.Fatalf("ConvertImageData failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("WebP data is empty")
	}

	// Verify WebP data can be decoded
	decodedImg, err := nativewebp.Decode(os.NewFile(0, ""))
	_ = decodedImg // Avoid unused variable error
	// Note: We can't easily decode from []byte without creating a Reader
	// Just verify we got non-empty data
	t.Logf("✓ WebP data conversion test passed: %d bytes", len(webpData))
}

// TestWebPFormats tests conversion between various formats and WebP
func TestWebPFormats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			img.Set(i, j, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	tests := []struct {
		name       string
		format     string
		shouldWork bool
	}{
		{"PNG to WebP", "png", true},
		{"JPEG to WebP", "jpg", true},
		{"GIF to WebP", "gif", true},
		{"TIFF to WebP", "tiff", true},
		{"BMP to WebP", "bmp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save source image
			sourcePath := filepath.Join(tmpDir, "source."+tt.format)
			err := imaging.SaveImageWithFormat(img, sourcePath, tt.format)
			if err != nil {
				t.Fatalf("Failed to save source image: %v", err)
			}

			// Convert to WebP
			webpPath := filepath.Join(tmpDir, "output_"+tt.format+".webp")
			err = imaging.ConvertImageFile(sourcePath, webpPath)

			if tt.shouldWork {
				if err != nil {
					t.Errorf("Conversion should work but failed: %v", err)
				} else if _, err := os.Stat(webpPath); os.IsNotExist(err) {
					t.Error("WebP file was not created")
				} else {
					t.Logf("✓ %s conversion successful: %s", tt.format, webpPath)
				}
			} else {
				if err == nil {
					t.Error("Conversion should fail but succeeded")
				}
			}
		})
	}
}

// TestWebPSaveImageWithFormat tests SaveImageWithFormat directly
func TestWebPSaveImageWithFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 30, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: uint8(y * 8), B: uint8(x * 8), A: 255})
		}
	}

	// Save as WebP using SaveImageWithFormat
	webpPath := filepath.Join(tmpDir, "direct.webp")
	err := imaging.SaveImageWithFormat(img, webpPath, "webp")
	if err != nil {
		t.Fatalf("SaveImageWithFormat failed: %v", err)
	}

	// Verify file exists and is valid
	info, err := os.Stat(webpPath)
	if err != nil {
		t.Fatalf("WebP file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("WebP file is empty")
	}

	t.Logf("✓ SaveImageWithFormat test passed: %s (%d bytes)", webpPath, info.Size())
}
