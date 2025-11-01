# Gimage - AI-Powered Image Generation and Processing

**Gimage** is a powerful tool for generating AI images and processing them with ease. Built with pure Go for maximum portability.

## 🚀 Two Ways to Use Gimage

### 1. Command-Line Tool (CLI)
A single binary with zero dependencies for local image operations.

### 2. Cloud API (AWS Lambda)
Production-ready serverless API for web applications and remote processing.

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

### 🔌 Integration Options
- **Claude Desktop**: Run as MCP server
- **Web Applications**: REST API via AWS Lambda
- **Serverless**: AWS Lambda on ARM64/Graviton2

## Quick Start - CLI

### 1. Installation

#### Homebrew (macOS/Linux) - Recommended

```bash
# Install gimage (tap is added automatically)
brew install apresai/tap/gimage
```

#### Upgrade via Homebrew

```bash
brew upgrade apresai/tap/gimage
```

#### Manual Installation

Download the latest release for your platform:

```bash
# macOS (Intel)
curl -L https://github.com/apresai/gimage/releases/latest/download/gimage-darwin-amd64 -o gimage
chmod +x gimage
sudo mv gimage /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/apresai/gimage/releases/latest/download/gimage-darwin-arm64 -o gimage
chmod +x gimage
sudo mv gimage /usr/local/bin/

# Linux
curl -L https://github.com/apresai/gimage/releases/latest/download/gimage-linux-amd64 -o gimage
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

## MCP Server for AI Assistants

Gimage can run as an MCP (Model Context Protocol) server, enabling AI assistants like Claude to generate and process images directly. The server communicates over stdio using the MCP protocol and exposes 10 tools for image generation and processing.

### Installation Methods

Choose the method that works best for you:

#### Method 1: Homebrew + MCP Server (Recommended for macOS/Linux)

This is the cleanest approach - install gimage CLI via Homebrew, then use it as an MCP server:

**Step 1: Install gimage via Homebrew**
```bash
# Install gimage (tap is added automatically)
brew install apresai/tap/gimage

# Verify installation
gimage --version
```

**Step 2: Configure Claude Desktop**

Edit your configuration file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "gimage": {
      "command": "gimage",
      "args": ["serve"]
    }
  }
}
```

**Benefits**:
- Uses the same gimage installation for both CLI and MCP
- Easy updates with `brew upgrade apresai/tap/gimage`
- No duplicate binaries

#### Method 2: npm Package (Alternative - Works on all platforms)

If you prefer npm or don't use Homebrew:

**Step 1: Install the npm package**
```bash
npm install -g @apresai/gimage-mcp
```

**Step 2: Configure Claude Desktop**

Edit your configuration file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "gimage": {
      "command": "npx",
      "args": ["-y", "@apresai/gimage-mcp"]
    }
  }
}
```

The npm package automatically downloads the correct gimage binary for your platform during installation.

**Note**: If you install via npm, you'll have a separate gimage binary just for MCP. If you also want the CLI, install via Homebrew.

### Setup Authentication

Before using the MCP server, configure your API credentials:

```bash
# For Gemini API (simplest, free tier available)
gimage auth gemini

# OR for Vertex AI
gimage auth vertex
```

The MCP server will automatically use credentials from:
- `~/.gimage/config.md` (created by auth commands)
- Environment variables (`GEMINI_API_KEY`, `VERTEX_API_KEY`, etc.)

### Start Using with Claude

After installation and authentication, restart Claude Desktop. You can then use natural language:

```
"Generate an image of a sunset over mountains"
"Resize photo.jpg to 800x600"
"Compress all images in the photos directory to 85% quality"
"Create a 2048x2048 photorealistic image of a wise old wizard"
```

### Available MCP Tools

The MCP server exposes 10 tools for AI assistants:

| Tool | Purpose |
|------|---------|
| `generate_image` | AI image generation from text |
| `resize_image` | Resize to specific dimensions |
| `scale_image` | Scale by factor (maintain aspect ratio) |
| `crop_image` | Crop to specific region |
| `compress_image` | Reduce file size with quality control |
| `convert_image` | Convert between formats |
| `batch_resize` | Resize multiple images concurrently |
| `batch_compress` | Compress multiple images |
| `batch_convert` | Convert multiple images to new format |
| `list_models` | List available AI models |

### Troubleshooting MCP Server

If the MCP server isn't working in Claude Desktop:

1. **Verify gimage is installed and in PATH:**
   ```bash
   which gimage
   gimage --version
   ```

2. **Test the serve command directly:**
   ```bash
   gimage serve --verbose
   ```
   This will show detailed logging to stderr. Press Ctrl+C to stop.

3. **Verify API credentials are configured:**
   ```bash
   gimage auth gemini  # or: gimage auth vertex
   ```

4. **Test image generation works outside MCP:**
   ```bash
   gimage generate "test image"
   ```

5. **Check Claude Desktop logs** for error messages:
   - **macOS**: `~/Library/Logs/Claude/`
   - **Linux**: `~/.config/Claude/logs/`

6. **Ensure config file exists and has correct permissions:**
   ```bash
   ls -la ~/.gimage/config.md
   ```
   Should be readable (permissions: `-rw-------` or `-rw-r--r--`)

### Environment Variables

The MCP server respects the same environment variables as the CLI:

```bash
export GEMINI_API_KEY="your-gemini-key"          # Gemini API
export VERTEX_API_KEY="your-vertex-key"          # Vertex AI Express Mode
export VERTEX_PROJECT="your-gcp-project"         # Vertex AI project
export VERTEX_LOCATION="us-central1"             # Vertex AI location
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
```

### MCP Documentation

For comprehensive guides and examples:

- **Usage Guide**: [docs/MCP_USAGE.md](docs/MCP_USAGE.md) - Complete setup and usage
- **Tool Reference**: [docs/MCP_TOOLS.md](docs/MCP_TOOLS.md) - Detailed tool documentation
- **Examples**: [docs/MCP_EXAMPLES.md](docs/MCP_EXAMPLES.md) - Real-world workflows
- **Implementation Plan**: [mcp.md](mcp.md) - Complete MCP architecture and implementation details

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

---

## Lambda API Distribution

Deploy Gimage as a serverless REST API on AWS Lambda for web application integration.

### Features

- **Serverless Architecture**: AWS Lambda on ARM64/Graviton2 (provided.al2023 runtime)
- **Auto-Scaling**: Handles 0 to thousands of requests automatically
- **Cost-Effective**: Pay only for what you use (~$0.25/month for 10K requests)
- **Global CDN**: S3 + CloudFront for fast image delivery
- **Production-Ready**: Full CORS, error handling, monitoring

### Quick Deploy

**Get deployed in under 1 hour!** See [QUICK_START_LAMBDA.md](QUICK_START_LAMBDA.md)

```bash
# Build Lambda function
make build-lambda

# Package for deployment
make package-lambda

# Deploy with CDK (requires infrastructure setup)
make deploy-lambda
```

Or follow the complete step-by-step guide in [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)

### API Endpoints

All operations available via REST API:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/generate` | POST | AI image generation |
| `/resize` | POST | Resize to dimensions |
| `/scale` | POST | Scale by factor |
| `/crop` | POST | Crop region |
| `/compress` | POST | Compress with quality |
| `/convert` | POST | Convert format |
| `/batch` | POST | Process multiple images |
| `/health` | GET | Health check |
| `/docs` | GET | **Interactive Swagger UI documentation** |
| `/openapi.yaml` | GET | OpenAPI specification |

**Try the API**: After deployment, visit `https://your-api-url/prod/docs` for interactive documentation!

### Integration Examples

**TypeScript/JavaScript:**
```typescript
const response = await fetch('https://your-api-url/generate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    prompt: 'a sunset over mountains',
    size: '1024x1024',
    response_format: 's3_url'
  })
});

const result = await response.json();
console.log('Image URL:', result.s3_url);
```

**Python:**
```python
import requests

response = requests.post('https://your-api-url/resize', json={
    'image': base64_image,
    'width': 800,
    'height': 600
})

result = response.json()
print(f"Resized: {result['width']}x{result['height']}")
```

**Go:**
```go
client := gimage.NewClient("https://your-api-url", apiKey)
result, err := client.GenerateImage(gimage.GenerateRequest{
    Prompt: "beautiful landscape",
    Size:   "1024x1024",
})
```

### Documentation

- **Quick Start**: [QUICK_START_LAMBDA.md](QUICK_START_LAMBDA.md) - Deploy in under 1 hour
- **Deployment Checklist**: [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md) - Step-by-step deployment
- **OpenAPI Spec**: [openapi.yaml](openapi.yaml) - Complete API reference
- **Integration Guide**: [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) - Client SDKs & examples
- **Deployment Plan**: [lambda.md](lambda.md) - Complete infrastructure setup
- **Status**: [LAMBDA_STATUS.md](LAMBDA_STATUS.md) - Implementation status

### Architecture

```
Client Request → API Gateway → Lambda (Go/ARM64) → {S3, Gemini, Vertex AI}
                                ↓
                         Small: base64 response
                         Large: S3 presigned URL
```

**Runtime**: AWS Lambda provided.al2023 (Amazon Linux 2023)
**Architecture**: ARM64 (Graviton2 processors)
**Package Size**: 17MB compressed, 42MB uncompressed
**Memory**: 2GB (configurable)
**Timeout**: 5 minutes max

---

## Building from Source

### CLI Binary

```bash
git clone https://github.com/apresai/gimage.git
cd gimage
make build
./bin/gimage --version
```

### Lambda Function

```bash
# Build for AWS Lambda ARM64
make build-lambda

# Package as deployment zip
make package-lambda

# Output: bin/lambda.zip (17MB)
```

## Requirements

### CLI
- **None!** Single static binary with zero dependencies
- Works on: macOS, Linux, Windows (x86_64, ARM64)
- No Python, no Node.js, no system libraries required

### Lambda API
- AWS Account
- AWS CLI configured
- Node.js 20+ (for CDK deployment)
- Go 1.22+ (for building)

## Support & Documentation

### CLI
- **Full Command Reference**: [COMMANDS.md](COMMANDS.md)
- **Configuration Guide**: See `gimage auth --help`

### Lambda API
- **OpenAPI Specification**: [openapi.yaml](openapi.yaml)
- **Integration Guide**: [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)
- **Deployment Guide**: [lambda.md](lambda.md)
- **Implementation Status**: [LAMBDA_STATUS.md](LAMBDA_STATUS.md)

### Community
- **GitHub Issues**: https://github.com/apresai/gimage/issues
- **Discussions**: https://github.com/apresai/gimage/discussions

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
