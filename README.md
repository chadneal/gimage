# Gimage - AI-Powered Image Generation and Processing CLI

**Gimage** is a powerful command-line tool for generating AI images and processing them with ease. Built with pure Go for maximum portability - a single binary with zero dependencies.

## What Can You Do with Gimage?

### 🎨 AI Image Generation
- Generate stunning images from text prompts using Google Gemini or Vertex AI
- Multiple AI models: Gemini 2.5 Flash, Imagen 3, Imagen 4
- Control size, style, and use negative prompts
- Reproducible results with seed values

### 🛠️ Image Processing
- **Resize** - Change image dimensions with high-quality resampling
- **Scale** - Scale images by factor (2x, 0.5x, etc.)
- **Crop** - Extract specific regions from images
- **Compress** - Reduce file size while maintaining quality
- **Convert** - Transform between formats (PNG, JPG, WebP, GIF, TIFF, BMP)

### ⚡ Batch Processing
- Process multiple images concurrently
- Configurable worker pool for optimal performance
- Apply any operation to entire directories

### 🔌 Claude Integration
- Run as an MCP server for Claude Desktop
- Expose all image operations as Claude tools

## Quick Start

### 1. Installation

Download the latest release for your platform:

```bash
# macOS
brew install gimage

# Or download from releases
curl -L https://github.com/chadneal/gimage/releases/latest/download/gimage-darwin-amd64 -o gimage
chmod +x gimage
sudo mv gimage /usr/local/bin/
```

### 2. Setup Authentication

#### For Gemini API (Simplest)
```bash
gimage auth gemini
```
Get your API key from: https://aistudio.google.com/app/apikey

#### For Vertex AI (3 Options)
```bash
gimage auth vertex
```
Choose from:
1. **Express Mode** - Simple API key (best for development)
2. **Service Account** - JSON credentials (best for production)
3. **Application Default Credentials** - Use your gcloud login

### 3. Generate Your First Image

```bash
gimage generate "a sunset over mountains with vibrant colors"
```

That's it! Your image will be saved as `generated_<timestamp>.png`

## Examples

### Image Generation

```bash
# Basic generation
gimage generate "futuristic city at night"

# Specify size and style
gimage generate "abstract art" --size 1024x1024 --style photorealistic

# Use Vertex AI Imagen
gimage generate "beautiful landscape" --api vertex --model imagen-4

# Use negative prompts to avoid unwanted elements
gimage generate "forest scene" --negative "people, buildings"

# Reproducible results with seed
gimage generate "random pattern" --seed 12345

# List all available models
gimage generate --list-models
```

### Image Processing

```bash
# Resize to specific dimensions
gimage resize photo.jpg 800 600

# Scale to 50% size
gimage scale photo.jpg 0.5

# Crop a region (x, y, width, height)
gimage crop photo.jpg 100 100 800 600

# Compress with custom quality
gimage compress photo.jpg --quality 85

# Convert format
gimage convert photo.png jpg
```

### Batch Processing

```bash
# Resize all images in a directory
gimage batch resize photos/ --width 800 --height 600 --output resized/

# Compress all images
gimage batch compress photos/ --quality 85 --output compressed/

# Convert all to WebP
gimage batch convert photos/ webp --output webp/

# Use 8 parallel workers for faster processing
gimage batch resize photos/ --width 1920 --height 1080 --workers 8
```

## Configuration

Gimage stores settings in `~/.gimage/config.md` (created automatically by `gimage auth` commands).

### Config File Format

```markdown
# Gimage Configuration

**gemini_api_key**: your-gemini-key
**vertex_api_key**: your-vertex-key
**vertex_project**: your-gcp-project
**vertex_location**: us-central1
**default_api**: gemini
**default_model**: gemini-2.5-flash-image
**default_size**: 1024x1024
**log_level**: info
```

### Priority Order
1. **Command-line flags** (highest)
2. **Environment variables** (`GEMINI_API_KEY`, `VERTEX_API_KEY`, etc.)
3. **Config file** (`~/.gimage/config.md`)
4. **Default values** (lowest)

### Environment Variables

```bash
export GEMINI_API_KEY="your-key"           # Gemini API key
export VERTEX_API_KEY="your-key"           # Vertex AI Express Mode key
export VERTEX_PROJECT="your-project"       # Vertex AI project ID
export VERTEX_LOCATION="us-central1"       # Vertex AI location
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"
```

## Available Models

### Gemini API (Google AI Studio)
- `gemini-2.5-flash-image` (default, recommended)
- `gemini-2.0-flash-preview-image-generation`

### Vertex AI
- `imagen-3.0-generate-002` (latest Imagen 3)
- `imagen-4` (newest, highest quality)

Run `gimage generate --list-models` to see all available models.

## Image Formats Supported

**Input & Output**: PNG, JPEG, WebP, GIF, TIFF, BMP

All processing operations preserve format by default, or you can specify output format with `--output` flag or use the `convert` command.

## Advanced Features

### Styles (Generation)
- `photorealistic` - Realistic photos
- `artistic` - Artistic interpretations
- `anime` - Anime/manga style

### Quality Settings
- Compression quality: 1-100 (default: 90)
- Resampling: Lanczos algorithm (highest quality)
- Transparency: Preserved for PNG, white background for JPEG

### Batch Processing
- Default: 4 parallel workers
- Configurable with `--workers` flag
- Processes entire directories
- Preserves directory structure with `--output`

## Use Cases

### For Designers & Artists
- Rapid prototyping of visual concepts
- Generate variations of ideas
- Quick image editing and optimization
- Batch resize for different platforms

### For Developers
- Generate placeholder images for apps
- Automate image processing pipelines
- Integrate AI generation into workflows
- Test with reproducible images (seeds)

### For Content Creators
- Create social media graphics
- Generate blog post illustrations
- Batch optimize images for web
- Convert formats for different platforms

### For AI/ML Engineers
- Generate training data
- Create synthetic datasets
- Automate image augmentation
- Integrate with Claude via MCP server

## MCP Server for Claude

Run gimage as an MCP server to use it directly in Claude Desktop:

```bash
# Add to Claude Desktop MCP config
{
  "mcpServers": {
    "gimage": {
      "command": "gimage",
      "args": ["serve"]
    }
  }
}
```

Claude can then:
- Generate images from prompts
- Resize, crop, compress images
- Batch process images
- Get image metadata

## Command Reference

| Command | Description |
|---------|-------------|
| `generate` | Generate images from text using AI |
| `resize` | Resize images to specific dimensions |
| `scale` | Scale images by factor |
| `crop` | Crop images to regions |
| `compress` | Compress images to reduce file size |
| `convert` | Convert between image formats |
| `batch` | Batch process multiple images |
| `auth` | Configure API credentials |
| `config` | Manage configuration |

Run `gimage [command] --help` for detailed usage.

## Building from Source

```bash
git clone https://github.com/chadneal/gimage.git
cd gimage
make build
./bin/gimage --version
```

## Requirements

- **None!** Single static binary with zero dependencies
- Works on: macOS, Linux, Windows (x86_64, ARM64)
- No Python, no Node.js, no system libraries required

## Support & Documentation

- **Full Command Reference**: See [COMMANDS.md](COMMANDS.md)
- **GitHub Issues**: https://github.com/chadneal/gimage/issues
- **Discussions**: https://github.com/chadneal/gimage/discussions

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Imaging](https://github.com/disintegration/imaging) - Image processing
- [Google Gen AI SDK](https://github.com/googleapis/go-genai) - Gemini API
- [Vertex AI SDK](https://cloud.google.com/go/vertexai) - Vertex AI

---

**Made with ❤️ for developers, designers, and AI enthusiasts**
