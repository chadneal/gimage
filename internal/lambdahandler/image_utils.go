package lambdahandler

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EncodeImageToBase64 encodes image bytes to base64 string
func EncodeImageToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64ToImage decodes a base64 string to image bytes
func DecodeBase64ToImage(base64Str string) ([]byte, error) {
	// Handle data URLs (e.g., "data:image/png;base64,...")
	if strings.HasPrefix(base64Str, "data:") {
		parts := strings.SplitN(base64Str, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		base64Str = parts[1]
	}

	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return data, nil
}

// IsBase64Input determines if the input string is base64 encoded or an S3 key
// Base64 strings are typically much longer and contain only base64 characters
func IsBase64Input(input string) bool {
	// Check for data URL prefix
	if strings.HasPrefix(input, "data:") {
		return true
	}

	// S3 keys are typically short paths (e.g., "images/abc123.png")
	// Base64 strings are long and contain only base64 chars
	if len(input) < 100 {
		return false // Likely an S3 key
	}

	// Check if it contains only base64 characters
	for _, c := range input {
		if !isBase64Char(c) {
			return false
		}
	}

	return true
}

// isBase64Char checks if a character is valid in base64
func isBase64Char(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '+' || c == '/' || c == '='
}

// LoadImageFromInput loads image data from either base64 string or S3 key
func LoadImageFromInput(ctx context.Context, s3Client *S3Client, input string) ([]byte, error) {
	if IsBase64Input(input) {
		// Decode base64
		return DecodeBase64ToImage(input)
	}

	// Download from S3
	return s3Client.Download(ctx, input)
}

// DetermineResponseFormat determines whether to return base64 or S3 URL based on image size
func DetermineResponseFormat(sizeBytes int64, requestedFormat string) string {
	// If user explicitly requested a format, use it
	if requestedFormat == "base64" || requestedFormat == "s3_url" {
		return requestedFormat
	}

	// Get max response size from env (default 512 KB)
	maxSizeKB := 512
	if maxSizeStr := os.Getenv("MAX_RESPONSE_SIZE_KB"); maxSizeStr != "" {
		if parsed, err := strconv.Atoi(maxSizeStr); err == nil {
			maxSizeKB = parsed
		}
	}

	maxSizeBytes := int64(maxSizeKB * 1024)

	// Use base64 for small images, S3 URL for large images
	if sizeBytes <= maxSizeBytes {
		return "base64"
	}

	return "s3_url"
}

// GenerateS3Key generates a unique S3 key for an image
func GenerateS3Key(format string) string {
	timestamp := time.Now().Unix()
	uuid := uuid.New().String()
	return fmt.Sprintf("images/%d-%s.%s", timestamp, uuid, format)
}

// GetContentType returns the MIME type for an image format
func GetContentType(format string) string {
	format = strings.ToLower(strings.TrimPrefix(format, "."))

	switch format {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "tiff", "tif":
		return "image/tiff"
	case "bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

// GetPresignedURLExpiration returns the expiration duration for presigned URLs
func GetPresignedURLExpiration() time.Duration {
	// Get expiration from env (default 60 minutes)
	expirationMinutes := 60
	if expStr := os.Getenv("PRESIGNED_URL_EXPIRATION_MINUTES"); expStr != "" {
		if parsed, err := strconv.Atoi(expStr); err == nil {
			expirationMinutes = parsed
		}
	}

	return time.Duration(expirationMinutes) * time.Minute
}
