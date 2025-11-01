# Generate Package

This package implements AI-powered image generation using the Google Gemini API, following IMAGE_CLI_PLAN.md PROMPT 3.

## Files

### gemini.go
Main Gemini API client implementation with the following features:
- **GeminiClient**: Handles interactions with Gemini API for image generation
- **NewGeminiClient**: Creates a new client with API key validation
- **GenerateImage**: Generates images from text prompts with retry logic (max 3 attempts)
- **ValidateCredentials**: Validates API credentials
- **SetModel**: Updates the model to use for generation

**Key Features**:
- Exponential backoff retry logic (1s initial, 10s max)
- Automatic error classification (retryable vs non-retryable)
- Support for multiple image generation models
- Configurable generation options (size, aspect ratio, style, negative prompts)
- Metadata extraction and storage

**Default Configuration**:
- Model: `gemini-2.5-flash-image`
- Size: `1024x1024`
- Max Retries: 3
- Initial Backoff: 1 second
- Max Backoff: 10 seconds

### prompt.go
Prompt enhancement and validation utilities:
- **EnhancePrompt**: Enhances user prompts for better AI generation results
- **EnhancePromptWithStyle**: Applies style templates (photorealistic, artistic, anime, etc.)
- **BuildPromptWithNegative**: Combines prompts with negative prompt guidance
- **ValidatePrompt**: Validates prompt length and content
- **ExtractKeywords**: Extracts keywords from prompts for metadata
- **TruncatePrompt**: Truncates prompts while preserving word boundaries
- **FormatPromptForDisplay**: Formats prompts for user-friendly display

**Style Templates**:
- `photorealistic`: Professional photography style
- `artistic`: Creative artistic interpretation
- `anime`: Anime/manga style
- `cinematic`: Cinematic composition
- `digital-art`: Digital illustration style
- `oil-painting`: Oil painting style
- `watercolor`: Watercolor painting style
- `3d-render`: 3D CGI rendering style

### download.go
Image saving and file management utilities:
- **SaveImage**: Saves generated images to disk
- **SaveImageWithMetadata**: Saves images with companion JSON metadata
- **GenerateOutputPath**: Generates timestamped output paths (`generated_YYYYMMDD_HHMMSS.{format}`)
- **GenerateOutputPathWithPrefix**: Custom prefix support
- **GenerateOutputPathInDir**: Directory-specific output paths
- **GenerateUniqueOutputPath**: Ensures unique filenames (auto-increments if exists)
- **ValidateOutputPath**: Validates output paths and permissions
- **EnsureOutputDir**: Creates output directories if needed
- **FileExists**: Checks file existence

**Path Format**: `generated_20251030_143022.png`

## Usage Examples

### Basic Image Generation
```go
import (
    "context"
    "github.com/apresai/gimage/internal/generate"
    "github.com/apresai/gimage/pkg/models"
)

// Create client
client, err := generate.NewGeminiClient("your-api-key")
if err != nil {
    log.Fatal(err)
}

// Generate image
ctx := context.Background()
options := models.GenerateOptions{
    Model: "gemini-2.5-flash-image",
    Size:  "1024x1024",
    Style: "photorealistic",
}

img, err := client.GenerateImage(ctx, "a beautiful sunset over mountains", options)
if err != nil {
    log.Fatal(err)
}

// Save image
outputPath := generate.GenerateOutputPath("png")
err = generate.SaveImage(img, outputPath)
if err != nil {
    log.Fatal(err)
}
```

### With Prompt Enhancement
```go
// Enhance prompt with style
prompt := generate.EnhancePromptWithStyle("a cat", "photorealistic")
// Result: "a cat, highly detailed photorealistic image, professional photography, 8k resolution, realistic lighting"

// Generate with enhanced prompt
img, err := client.GenerateImage(ctx, prompt, options)
```

### With Negative Prompts
```go
options := models.GenerateOptions{
    Model:          "gemini-2.5-flash-image",
    Size:           "1024x1024",
    NegativePrompt: "blurry, low quality, distorted",
}

img, err := client.GenerateImage(ctx, "portrait of a person", options)
```

### With Metadata
```go
// Save image with metadata JSON file
err = generate.SaveImageWithMetadata(img, outputPath)
// Creates: generated_20251030_143022.png
//          generated_20251030_143022.png.json
```

### Unique Path Generation
```go
// Automatically handles existing files
path1 := generate.GenerateUniqueOutputPath("png")
path2 := generate.GenerateUniqueOutputPath("png")
// path1: generated_20251030_143022.png
// path2: generated_20251030_143022_1.png (if path1 exists)
```

## Error Handling

The package implements comprehensive error handling:

### Retryable Errors
Automatically retried with exponential backoff:
- Rate limit errors (429)
- Quota exceeded
- Timeout errors
- Connection errors
- Service unavailable (503)
- Deadline exceeded

### Non-Retryable Errors
Return immediately without retry:
- Invalid API key
- Invalid prompts
- Permission errors
- Invalid parameters

### Error Examples
```go
// Empty prompt
_, err := client.GenerateImage(ctx, "", options)
// Error: "prompt cannot be empty"

// Invalid credentials
err := client.ValidateCredentials()
// Error: "credential validation failed: ..."

// API failure with retry
_, err := client.GenerateImage(ctx, prompt, options)
// Error: "failed after 3 attempts: ..."
```

## Configuration

### Generation Options
```go
type GenerateOptions struct {
    Model          string  // "gemini-2.5-flash-image", etc.
    Size           string  // "1024x1024", "512x768", etc.
    AspectRatio    string  // "1:1", "16:9", "9:16", etc.
    Style          string  // Style template name
    NegativePrompt string  // What to avoid in generation
    Seed           int64   // Random seed for reproducibility
}
```

### Default Values
- Model: `gemini-2.5-flash-image`
- Size: `1024x1024`
- Output Format: `png`
- Output Prefix: `generated`
- File Permissions: `0644`
- Directory Permissions: `0755`

## Testing

The package includes comprehensive table-driven tests:

```bash
# Run all tests
go test ./internal/generate/... -v

# Run with coverage
go test ./internal/generate/... -cover

# Run specific test
go test ./internal/generate/... -run TestGenerateImage
```

**Test Coverage**:
- Unit tests for all public functions
- Error case testing
- Edge case handling
- File I/O testing with temporary directories
- Prompt enhancement validation
- Path generation testing

## Dependencies

- `google.golang.org/genai` - Google Gemini API SDK
- Standard library only (no external image processing dependencies)

## Implementation Notes

### Pure Go Implementation
- Zero C dependencies
- Uses `google.golang.org/genai` SDK
- All string manipulation using pure Go
- No external image processing libraries required in this package

### Retry Logic
- Max 3 attempts
- Initial backoff: 1 second
- Backoff multiplier: 2x
- Max backoff: 10 seconds
- Automatic error classification

### File Naming Convention
Format: `{prefix}_{YYYYMMDD}_{HHMMSS}.{format}`
- Prefix: Default `generated`, configurable
- Date: Year, month, day
- Time: Hour, minute, second
- Format: Image format extension (png, jpg, webp, etc.)

### Metadata Schema
Generated JSON metadata includes:
```json
{
  "model": "gemini-2.5-flash-image",
  "prompt": "enhanced prompt text",
  "size": "1024x1024",
  "style": "photorealistic",
  "generated": "2025-10-30T14:30:22Z",
  "api": "gemini",
  "seed": "12345",
  "negative_prompt": "blurry, low quality",
  "format": "png",
  "width": 1024,
  "height": 1024,
  "size_bytes": 123456
}
```

## Future Enhancements

Potential improvements for future iterations:
- Support for batch image generation
- Image-to-image transformation
- Style transfer capabilities
- Advanced prompt templating
- Cost estimation before generation
- Progress callbacks for long operations
- Caching layer for repeated prompts
- Image quality assessment
