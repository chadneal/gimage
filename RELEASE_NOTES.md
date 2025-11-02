# Release v1.0 - Initial Public Release

**Release Date**: November 1, 2025

We're excited to announce the first public release of **gimage** - a powerful, pure-Go CLI tool and MCP server for AI-powered image generation and processing!

---

## 🎉 What is Gimage?

Gimage is a versatile image generation and processing tool that can be used in two powerful ways:

### 1. **Command-Line Tool (CLI)**
A single, zero-dependency binary for local image operations on macOS, Linux, and Windows.

### 2. **MCP Server for AI Assistants**
Integration with Claude Desktop and other AI assistants via the Model Context Protocol, enabling natural language image operations.

---

## ✨ Major Features

### 🎨 AI Image Generation

Generate stunning images from text prompts using state-of-the-art AI models:

- **Multiple AI Backends**:
  - Google Gemini API (Gemini 2.5 Flash Image - recommended)
  - Vertex AI (Imagen 3, Imagen 4 - highest quality)
  - Auto-detection based on model selection

- **Flexible Configuration**:
  - Multiple sizes: 256x256 to 2048x2048
  - Style controls: photorealistic, artistic, anime
  - Negative prompts to exclude unwanted elements
  - Seed-based reproducible generation

- **Multi-Backend Architecture**:
  - Gemini REST API client
  - Vertex AI REST client (Express Mode with API key)
  - Vertex AI SDK client (Full Mode with service accounts)

**Example**:
```bash
# Generate with Gemini
gimage generate "a sunset over mountains with vibrant colors"

# Use Vertex AI Imagen 4 for highest quality
gimage generate "beautiful landscape" --api vertex --model imagen-4 --size 2048x2048

# Reproducible generation with seed
gimage generate "abstract patterns" --seed 42
```

### 🛠️ Comprehensive Image Processing

High-quality image processing operations with pure Go (zero C dependencies):

| Operation | Description | Example |
|-----------|-------------|---------|
| **Resize** | Change to exact dimensions | `gimage resize photo.jpg 800 600` |
| **Scale** | Scale by factor (preserves aspect ratio) | `gimage scale photo.jpg 0.5` |
| **Crop** | Extract specific region | `gimage crop photo.jpg 100 100 800 600` |
| **Compress** | Reduce file size with quality control | `gimage compress photo.jpg --quality 85` |
| **Convert** | Transform between formats | `gimage convert photo.png jpg` |

**Supported Formats**: PNG, JPEG, WebP, GIF, TIFF, BMP

**Quality Features**:
- Lanczos resampling (highest quality resizing)
- Transparency preservation for PNG
- Configurable JPEG quality (1-100)

### ⚡ Batch Processing

Process entire directories with concurrent workers:

```bash
# Resize all images in a directory
gimage batch resize photos/ --width 1920 --height 1080 --output resized/

# Compress all images with 8 parallel workers
gimage batch compress photos/ --quality 85 --workers 8 --output compressed/

# Convert all images to WebP for web optimization
gimage batch convert photos/ webp --output webp/
```

**Performance**:
- Configurable worker pools (default: CPU cores, max: 16)
- Concurrent processing for fast batch operations
- Progress tracking with success/failure counts

### 🔌 MCP Server Integration

**Model Context Protocol (MCP)** server for seamless AI assistant integration:

**10 Tools Exposed**:
1. `generate_image` - AI image generation from text
2. `resize_image` - Resize to specific dimensions
3. `scale_image` - Scale by factor
4. `crop_image` - Crop to region
5. `compress_image` - Reduce file size
6. `convert_image` - Convert formats
7. `batch_resize` - Batch resize operations
8. `batch_compress` - Batch compression
9. `batch_convert` - Batch format conversion
10. `list_models` - List available AI models

**Protocol Implementation**:
- JSON-RPC 2.0 over STDIO transport
- Proper notification vs request handling
- Structured error responses with `isError` flag
- Token-efficient tool descriptions
- Comprehensive input validation with JSON schemas

**Claude Desktop Integration**:
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

**Natural Language Usage**:
```
"Generate an image of a sunset over mountains"
"Resize photo.jpg to 800x600"
"Compress all images in the photos directory to 85% quality"
```

### 🔐 Authentication & Configuration

**Interactive Authentication Setup**:
```bash
# Gemini API (simplest - free tier available)
gimage auth gemini

# Vertex AI (3 modes: Express, Service Account, ADC)
gimage auth vertex
```

**Configuration File** (`~/.gimage/config.md`):
- Markdown format for readability
- Secure permissions (0600 automatically set)
- Environment variable overrides supported

**Priority Order**:
1. Command-line flags (highest)
2. Environment variables
3. Config file
4. Default values (lowest)

### 📦 Distribution Channels

**Three Installation Methods**:

1. **Homebrew** (macOS/Linux - Recommended):
   ```bash
   # Tap is added automatically
   brew install apresai/tap/gimage
   ```

2. **npm Package** (Cross-platform MCP server):
   ```bash
   npm install -g @apresai/gimage-mcp
   ```

3. **Manual Download** (Direct binary):
   - Download from GitHub Releases
   - Single binary, no dependencies

**Supported Platforms**:
- macOS (Intel x86_64, Apple Silicon ARM64)
- Linux (x86_64, ARM64)
- Windows (x86_64)

### 🏗️ AWS Lambda Support (Optional)

**Serverless REST API** deployment option:

- **Runtime**: AWS Lambda provided.al2023 (Amazon Linux 2023)
- **Architecture**: ARM64/Graviton2 processors
- **Auto-scaling**: 0 to thousands of requests
- **Cost-effective**: ~$0.25/month for 10K requests
- **Global CDN**: S3 + CloudFront integration

**API Endpoints**: `/generate`, `/resize`, `/scale`, `/crop`, `/compress`, `/convert`, `/batch`, `/health`, `/docs`

---

## 🎯 Design Philosophy

### Pure Go Implementation
- **Zero C dependencies** - Single static binary
- **Maximum portability** - Works anywhere Go runs
- **Cross-compilation** - Build for any platform from any platform
- **No system libraries** - No apt, brew, or system package requirements

### MCP Best Practices
- **Workflow-centric tool design** - Tools accomplish complete tasks
- **Three-tier error handling** - Transport, protocol, and application errors
- **Token-efficient descriptions** - Optimized for LLM context windows
- **Explicit input validation** - Comprehensive JSON schemas
- **Proper STDIO hygiene** - All logging to stderr, no stdout pollution

### User Experience
- **Intelligent path validation** - Automatic fallback to writable directories
- **Clear error messages** - Actionable guidance for troubleshooting
- **Structured responses** - Consistent success/failure/warning format
- **Graceful degradation** - Works with partial configuration

---

## 📚 Documentation

Comprehensive documentation included:

### General Documentation
- **README.md** - Complete overview and quick start guide
- **COMMANDS.md** - Full CLI command reference

### MCP Server Documentation
- **docs/MCP_USAGE.md** - Setup and usage guide
- **docs/MCP_TOOLS.md** - Detailed tool reference (all 10 tools)
- **docs/MCP_EXAMPLES.md** - Real-world workflow examples

### Lambda/API Documentation
- **QUICK_START_LAMBDA.md** - Deploy in under 1 hour
- **DEPLOYMENT_CHECKLIST.md** - Step-by-step deployment
- **INTEGRATION_GUIDE.md** - Client SDKs and examples
- **openapi.yaml** - Complete OpenAPI 3.0 specification

### Architecture Documentation
- **mcp.md** - MCP implementation architecture
- **lambda.md** - Lambda deployment architecture
- **IMAGE_CLI_PLAN.md** - Development roadmap

**Live Documentation**:
- Interactive Swagger UI at `/docs` endpoint (Lambda deployments)
- GitHub README: https://github.com/apresai/gimage

---

## 🔧 Technical Highlights

### MCP Protocol Implementation
- **Protocol Version**: 2024-11-05
- **Transport**: STDIO (standard input/output)
- **Format**: JSON-RPC 2.0
- **Capabilities**: Tools (with 10 exposed operations)
- **Notification Handling**: Proper detection and no-response behavior
- **Error Codes**: Standard JSON-RPC error codes (-32700 to -32603)

### Image Processing
- **Library**: `github.com/disintegration/imaging` (pure Go)
- **Resampling**: Lanczos algorithm (highest quality)
- **Concurrency**: Worker pool pattern with configurable parallelism
- **Memory Efficiency**: Streaming processing for large images

### API Integration
- **Gemini SDK**: `github.com/googleapis/go-genai`
- **Vertex SDK**: `cloud.google.com/go/vertexai`
- **HTTP Client**: Custom REST clients with retry logic
- **Authentication**: OAuth 2.0, API keys, service accounts, ADC

### Build System
- **Framework**: Makefile-based build automation
- **Testing**: Comprehensive unit tests for all tools
- **Cross-compilation**: Build for all platforms simultaneously
- **Packaging**: Automated Lambda zip creation

---

## 🚀 Getting Started

### Quick Start (5 minutes)

**1. Install via Homebrew**:
```bash
# Tap is added automatically
brew install apresai/tap/gimage
```

**2. Configure Authentication**:
```bash
gimage auth gemini
# Get free API key: https://aistudio.google.com/app/apikey
```

**3. Generate Your First Image**:
```bash
gimage generate "a sunset over mountains with vibrant colors"
```

**4. Process Images**:
```bash
gimage resize photo.jpg 800 600
gimage compress photo.jpg --quality 85
```

### MCP Server Setup (Claude Desktop)

**1. Install gimage** (via Homebrew or npm)

**2. Configure Claude Desktop**:

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:
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

**3. Setup Authentication**:
```bash
gimage auth gemini
```

**4. Restart Claude Desktop** and start using natural language!

---

## 🎓 Use Cases

### For Designers & Artists
- Rapid prototyping of visual concepts
- Generate variations of design ideas
- Batch resize for different platforms
- Quick image editing and optimization

### For Developers
- Generate placeholder images for apps
- Automate image processing pipelines
- Test with reproducible images (seeds)
- Integrate AI generation into workflows

### For Content Creators
- Create social media graphics
- Generate blog post illustrations
- Batch optimize images for web
- Convert formats for different platforms

### For AI/ML Engineers
- Generate synthetic training data
- Create image datasets at scale
- Integrate with Claude via MCP
- Automate image augmentation pipelines

---

## 📊 Performance Benchmarks

### Single Image Operations
- **Resize/Scale/Crop**: < 1 second (typical 3000x2000 image)
- **Compress**: 1-3 seconds (depending on size and quality)
- **Format Convert**: < 1 second
- **AI Generation (Gemini)**: 5-15 seconds (network-bound)
- **AI Generation (Vertex Imagen 4)**: 10-30 seconds (quality vs speed)

### Batch Operations (100 images, 4 workers)
- **Batch Resize**: ~10-30 seconds
- **Batch Compress**: ~20-60 seconds
- **Batch Convert**: ~15-45 seconds

---

## 🔒 Security

### Current Security Posture
✅ STDIO-only transport (no network exposure for MCP)
✅ Path validation prevents directory traversal
✅ No shell command execution (pure Go)
✅ API keys stored in secure config (0600 permissions)
✅ No SQL injection risk (no database)
✅ Input sanitization for file paths

### Best Practices
- Config file automatically created with owner-only permissions
- Environment variables supported for CI/CD
- Tilde expansion for home directory paths
- Writable directory validation with fallbacks

---

## 🐛 Known Limitations

1. **MCP Server**: Currently STDIO transport only (no HTTP/SSE endpoints)
2. **Testing**: Integration tests for MCP protocol not yet implemented
3. **Retry Logic**: No circuit breaker pattern for API calls yet
4. **Logging**: Unstructured logging (structured logging planned)
5. **Caching**: No semantic caching for repeated operations

See **roadmap.md** for planned improvements and future features.

---

## 🛣️ Roadmap

### Critical Priority (Next 1-2 months)
- Automated versioning and release pipeline
- MCP integration tests
- Circuit breaker pattern for API calls

### High Priority (3-6 months)
- Structured logging (zerolog/zap)
- Background removal with AI
- Smart crop with face detection
- Tool annotations (MCP 2025-06-18 spec)

### Future Features (Community Proposals)
- Image upscaling with AI enhancement
- Batch watermarking
- Metadata editor (EXIF/IPTC)
- Image similarity search
- Collage/mosaic generation
- Color palette extraction
- Animated GIF/video creation
- OCR and text extraction

See **roadmap.md** for complete analysis and feature proposals.

---

## 🙏 Credits

Built with these excellent open-source libraries:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Imaging](https://github.com/disintegration/imaging) - Pure Go image processing
- [Google Gen AI SDK](https://github.com/googleapis/go-genai) - Gemini API client
- [Vertex AI SDK](https://cloud.google.com/go/vertexai) - Vertex AI integration

---

## 📖 Resources

### Documentation
- **GitHub Repository**: https://github.com/apresai/gimage
- **README**: https://github.com/apresai/gimage#readme
- **MCP Tools Reference**: https://github.com/apresai/gimage/blob/main/docs/MCP_TOOLS.md
- **Installation Guide**: https://github.com/apresai/gimage#installation
- **Authentication Setup**: https://github.com/apresai/gimage#setup-authentication

### Distribution
- **Homebrew Tap**: https://github.com/apresai/homebrew-tap
- **npm Package**: https://www.npmjs.com/package/@apresai/gimage-mcp
- **GitHub Releases**: https://github.com/apresai/gimage/releases

### Community
- **Issues**: https://github.com/apresai/gimage/issues
- **Discussions**: https://github.com/apresai/gimage/discussions
- **Pull Requests**: https://github.com/apresai/gimage/pulls

---

## 📝 License

MIT License - see [LICENSE](https://github.com/apresai/gimage/blob/main/LICENSE) file for details.

---

## 🎯 Summary

Gimage v1.0 provides a comprehensive, production-ready solution for AI-powered image generation and processing. With its pure-Go implementation, multi-backend architecture, and seamless MCP integration, it serves developers, designers, content creators, and AI engineers across multiple platforms and use cases.

**Key Differentiators**:
- ✅ Zero dependencies - single binary distribution
- ✅ Multiple AI backends - Gemini and Vertex AI
- ✅ MCP server - natural language AI assistant integration
- ✅ Comprehensive tooling - 10 MCP tools + CLI commands
- ✅ Production-ready - proper error handling, validation, documentation
- ✅ Cross-platform - macOS, Linux, Windows support

**Get Started Today**:
```bash
# Install
# Tap is added automatically && brew install apresai/tap/gimage

# Authenticate
gimage auth gemini

# Generate
gimage generate "your creative prompt here"
```

---

**Made with ❤️ for developers, designers, and AI enthusiasts**

**Version**: 1.0
**Release Date**: November 1, 2025
**Maintainer**: Chad Neal
**Repository**: https://github.com/apresai/gimage
