# Gimage MCP Tools Reference

Complete reference for all 10 MCP tools available in the gimage server.

## Tool Index

1. [generate_image](#generate_image) - AI image generation
2. [resize_image](#resize_image) - Resize to dimensions
3. [scale_image](#scale_image) - Scale by factor
4. [crop_image](#crop_image) - Crop to region
5. [compress_image](#compress_image) - Compress file size
6. [convert_image](#convert_image) - Convert formats
7. [batch_resize](#batch_resize) - Batch resize
8. [batch_compress](#batch_compress) - Batch compress
9. [batch_convert](#batch_convert) - Batch convert
10. [list_models](#list_models) - List AI models

---

## generate_image

Generate an AI image from a text prompt using Gemini or Vertex AI.

### Description

Creates images from text descriptions using state-of-the-art AI models. Supports multiple models (Gemini 2.5 Flash, Imagen 3, Imagen 4), various sizes up to 2048x2048, and style controls. Can use negative prompts to exclude unwanted elements and seeds for reproducible generation.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `prompt` | string | Yes | - | Text description of the image to generate |
| `output` | string | No | Auto-generated | Output file path |
| `size` | string | No | "1024x1024" | Image dimensions |
| `model` | string | No | "gemini-2.5-flash-image" | AI model to use |
| `style` | string | No | - | Image style (photorealistic, artistic, anime) |
| `negative` | string | No | - | Negative prompt (what to exclude) |
| `seed` | integer | No | - | Random seed for reproducibility |

### Supported Sizes

- `256x256`
- `512x512`
- `1024x1024` (default)
- `1024x1792`
- `1792x1024`
- `2048x2048` (Vertex AI only)

### Supported Models

- **gemini-2.5-flash-image** (default, recommended)
- **gemini-2.0-flash-preview-image-generation**
- **imagen-3.0-generate-002** (requires Vertex AI)
- **imagen-4** (requires Vertex AI, highest quality)

### Returns

```json
{
  "success": true,
  "output_path": "/absolute/path/to/generated_1234567890.png",
  "size": "1024x1024",
  "model": "gemini-2.5-flash-image",
  "prompt": "a sunset over mountains"
}
```

### Examples

**Basic generation:**
```
Generate an image of a sunset over mountains
```

**With style:**
```
Create a photorealistic image of a wise old wizard
```

**With size and negative prompt:**
```
Generate a 1024x1792 image of a forest scene, but exclude any people or buildings
```

**Reproducible generation:**
```
Generate an image with seed 42 of abstract patterns
```

---

## resize_image

Resize an image to specific dimensions.

### Description

Resizes an image to exact width and height using high-quality Lanczos resampling. Note: Aspect ratio is NOT preserved unless dimensions match original ratio. Use `scale_image` if you want to maintain aspect ratio.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `input` | string | Yes | Input image file path |
| `width` | integer | Yes | Target width in pixels (minimum: 1) |
| `height` | integer | Yes | Target height in pixels (minimum: 1) |
| `output` | string | No | Output file path (default: auto-generated) |

### Returns

```json
{
  "success": true,
  "output_path": "/absolute/path/to/photo_resized.jpg",
  "original_size": "3000x2000",
  "new_size": "800x600"
}
```

### Examples

```
Resize photo.jpg to 800x600 pixels
Resize landscape.png to 1920x1080 and save as web-version.png
```

---

## scale_image

Scale an image by a factor while preserving aspect ratio.

### Description

Scales an image proportionally by a multiplication factor. Use this when you want to make an image larger or smaller while maintaining its aspect ratio. Uses high-quality Lanczos resampling.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `input` | string | Yes | Input image file path |
| `factor` | number | Yes | Scale factor (0.1 to 10.0) |
| `output` | string | No | Output file path (default: auto-generated) |

### Scale Factor Examples

- `0.5` = Half size
- `0.25` = Quarter size
- `2.0` = Double size
- `1.5` = 50% larger

### Returns

```json
{
  "success": true,
  "output_path": "/absolute/path/to/photo_scaled.jpg",
  "scale_factor": 0.5,
  "original_size": "2000x1500",
  "new_size": "1000x750"
}
```

### Examples

```
Scale photo.jpg to half its size
Make image.png twice as large (factor 2.0)
Reduce dimensions by 25% (factor 0.75)
```

---

## crop_image

Crop an image to a specific rectangular region.

### Description

Extracts a rectangular region from an image. Specify the top-left corner coordinates (x, y) and the width and height of the region. Useful for removing borders, focusing on specific areas, or extracting thumbnails.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `input` | string | Yes | Input image file path |
| `x` | integer | Yes | X coordinate of top-left corner (0 = left edge) |
| `y` | integer | Yes | Y coordinate of top-left corner (0 = top edge) |
| `width` | integer | Yes | Width of crop region in pixels (minimum: 1) |
| `height` | integer | Yes | Height of crop region in pixels (minimum: 1) |
| `output` | string | No | Output file path (default: auto-generated) |

### Returns

```json
{
  "success": true,
  "output_path": "/absolute/path/to/photo_cropped.jpg",
  "crop_region": "(100,100,800,600)",
  "crop_size": "800x600"
}
```

### Examples

```
Crop photo.jpg starting at (100, 100) with width 800 and height 600
Extract a 500x500 square from the top-left corner of image.png
```

---

## compress_image

Compress an image to reduce file size.

### Description

Reduces image file size while maintaining visual quality. Quality ranges from 1 (lowest quality, smallest file) to 100 (highest quality, largest file). Default is 90 which provides excellent quality with good compression. Most effective on JPEG images. PNG images are compressed losslessly.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `input` | string | Yes | - | Input image file path |
| `quality` | integer | No | 90 | Compression quality (1-100) |
| `output` | string | No | Auto-generated | Output file path |

### Recommended Quality Settings

- **95-100**: Archival quality, minimal compression
- **90**: Recommended for web (default)
- **85**: Good for mobile devices
- **75**: Acceptable for thumbnails
- **60-70**: Heavy compression, visible quality loss

### Returns

```json
{
  "success": true,
  "output_path": "/absolute/path/to/photo_compressed.jpg",
  "quality": 85,
  "original_size_bytes": 2500000,
  "compressed_size_bytes": 450000,
  "compression_ratio": "0.18",
  "savings_bytes": 2050000,
  "savings_percent": "82.0%",
  "original_size_human": "2.4 MB",
  "compressed_size_human": "439.5 KB"
}
```

### Examples

```
Compress photo.jpg to 85% quality
Reduce file size of large-image.png
Compress with 75% quality for thumbnails
```

---

## convert_image

Convert an image between different formats.

### Description

Converts images between PNG, JPG/JPEG, WebP, GIF, TIFF, and BMP formats. Useful for web optimization (converting to WebP), compatibility (PNG to JPG), or specific application requirements. Format detection is automatic based on file extension.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `input` | string | Yes | Input image file path |
| `format` | string | Yes | Target format (png, jpg, jpeg, webp, gif, tiff, bmp) |
| `output` | string | No | Output file path (default: auto-generated with new extension) |

### Supported Formats

- **PNG**: Lossless, supports transparency
- **JPG/JPEG**: Lossy, best for photos
- **WebP**: Modern format, great compression
- **GIF**: Animated images, limited colors
- **TIFF**: High-quality, large files
- **BMP**: Uncompressed, very large files

### Returns

```json
{
  "success": true,
  "output_path": "/absolute/path/to/image.webp",
  "original_format": "png",
  "new_format": "webp",
  "original_size": "1.2 MB",
  "new_size": "245.3 KB"
}
```

### Examples

```
Convert photo.png to JPEG format
Change image.jpg to WebP for better web performance
Convert screenshot.bmp to PNG
```

---

## batch_resize

Resize multiple images concurrently.

### Description

Processes all image files (PNG, JPG, WebP, GIF, TIFF, BMP) in a directory and resizes them to specified dimensions. Uses parallel workers for fast processing of large batches.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `input_dir` | string | Yes | - | Input directory containing images |
| `width` | integer | Yes | - | Target width in pixels (minimum: 1) |
| `height` | integer | Yes | - | Target height in pixels (minimum: 1) |
| `output_dir` | string | Yes | - | Output directory (created if doesn't exist) |
| `workers` | integer | No | CPU cores | Number of parallel workers (1-16) |

### Returns

```json
{
  "success": true,
  "processed": 45,
  "failed": 2,
  "total": 47,
  "output_dir": "/absolute/path/to/output",
  "errors": [
    "corrupted.jpg: failed to decode image",
    "locked.png: permission denied"
  ]
}
```

### Examples

```
Resize all images in vacation-photos folder to 1920x1080, save to resized-photos
Batch resize images in products/ to 600x600 using 8 workers
```

---

## batch_compress

Compress multiple images concurrently.

### Description

Processes all image files in a directory with specified quality setting to reduce file sizes. Reports total space saved across all images. Uses parallel workers for efficient processing.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `input_dir` | string | Yes | - | Input directory containing images |
| `quality` | integer | No | 85 | Compression quality (1-100) |
| `output_dir` | string | Yes | - | Output directory (created if doesn't exist) |
| `workers` | integer | No | CPU cores | Number of parallel workers (1-16) |

### Returns

```json
{
  "success": true,
  "processed": 50,
  "failed": 0,
  "total": 50,
  "output_dir": "/absolute/path/to/compressed",
  "total_original_size": "125.5 MB",
  "total_new_size": "28.3 MB",
  "total_savings": "97.2 MB",
  "savings_percent": "77.5%"
}
```

### Examples

```
Compress all images in photos/ to 85% quality, save to compressed/
Batch compress with 90% quality using 8 workers
```

---

## batch_convert

Convert multiple images to a different format concurrently.

### Description

Converts all image files in a directory to a specified format. Useful for converting entire directories to WebP for web optimization, or to PNG for lossless archival. Maintains original filenames with new extensions.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `input_dir` | string | Yes | - | Input directory containing images |
| `format` | string | Yes | - | Target format (png, jpg, jpeg, webp, gif, tiff, bmp) |
| `output_dir` | string | Yes | - | Output directory (created if doesn't exist) |
| `workers` | integer | No | CPU cores | Number of parallel workers (1-16) |

### Returns

```json
{
  "success": true,
  "processed": 30,
  "failed": 0,
  "total": 30,
  "output_dir": "/absolute/path/to/webp-images"
}
```

### Examples

```
Convert all images in photos/ to WebP format, save to webp-images/
Batch convert to PNG using 8 workers for faster processing
```

---

## list_models

List all available AI image generation models.

### Description

Returns detailed information about all available AI image generation models, including their capabilities, providers, maximum resolutions, and authentication requirements. Use this to discover which models are available before generating images.

### Parameters

None

### Returns

```json
{
  "models": [
    {
      "name": "gemini-2.5-flash-image",
      "provider": "Google Gemini API",
      "description": "Latest Gemini 2.5 Flash model...",
      "max_resolution": "1024x1792 or 1792x1024",
      "requires_api_key": true,
      "api_key_env": "GEMINI_API_KEY",
      "supports_styles": true,
      "supports_negative": true,
      "supports_seed": true,
      "free_tier": true,
      "free_tier_limit": "1500 requests/day"
    },
    // ... more models
  ],
  "total": 4
}
```

### Examples

```
List all available AI models
Show me what image generation models I can use
What models support 2K resolution?
```

---

## Error Handling

All tools return errors in a consistent format:

```json
{
  "error": {
    "code": -32603,
    "message": "Tool execution failed: file not found: /path/to/missing.jpg"
  }
}
```

### Common Error Codes

- **-32602**: Invalid parameters (missing required field, invalid type, out of range)
- **-32603**: Execution error (file not found, permission denied, API error)
- **-32601**: Method not found (invalid tool name)

### Error Messages

Error messages are designed to be clear and actionable:

- ✅ "Failed to open image: file not found: /path/to/photo.jpg"
- ✅ "Crop region (100,100,2000,1500) extends beyond image bounds (1000x800)"
- ✅ "Quality must be between 1 and 100, got: 150"

---

## Performance Notes

### Single Image Operations

- **Fast**: resize, scale, crop, convert (< 1 second for typical images)
- **Medium**: compress (1-3 seconds depending on size and quality)
- **Slow**: generate (5-30 seconds depending on model and size)

### Batch Operations

- Uses parallel workers (default: number of CPU cores)
- Processing 100 images:
  - Resize: ~10-30 seconds (4 workers)
  - Compress: ~20-60 seconds (4 workers)
  - Convert: ~15-45 seconds (4 workers)

### Tips for Better Performance

1. Use batch operations for multiple images instead of repeated single operations
2. Increase workers for faster batch processing (up to 16)
3. Use smaller images when possible
4. For generation, use Gemini 2.5 Flash for fastest results

---

## See Also

- [Usage Guide](MCP_USAGE.md) - Complete setup and usage instructions
- [Examples](MCP_EXAMPLES.md) - Real-world usage examples
- [Main Documentation](../README.md) - Full project documentation
