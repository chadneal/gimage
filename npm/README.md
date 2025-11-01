# @chadneal/gimage-mcp

MCP (Model Context Protocol) server for AI-powered image generation and processing with gimage.

## Quick Start

### Installation

```bash
npm install -g @chadneal/gimage-mcp
```

Or use without installation:

```bash
npx @chadneal/gimage-mcp
```

### Configure Claude Desktop

Add to your Claude Desktop MCP configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "gimage": {
      "command": "npx",
      "args": ["-y", "@chadneal/gimage-mcp"]
    }
  }
}
```

### Setup Authentication

Before using, configure your Gemini API key:

```bash
gimage auth gemini
```

Get your free API key from: https://aistudio.google.com/app/apikey

### Restart Claude Desktop

Quit and reopen Claude Desktop to load the MCP server.

## Features

The MCP server provides 10 tools for AI assistants:

### Image Generation
- **generate_image** - Create AI images from text prompts
  - Multiple models: Gemini 2.5 Flash, Imagen 3, Imagen 4
  - Sizes up to 2048x2048
  - Style controls (photorealistic, artistic, anime)
  - Negative prompts
  - Reproducible generation with seeds

### Image Processing
- **resize_image** - Resize to specific dimensions
- **scale_image** - Scale by factor (maintain aspect ratio)
- **crop_image** - Crop to specific region
- **compress_image** - Reduce file size with quality control
- **convert_image** - Convert between formats (PNG, JPG, WebP, GIF, TIFF, BMP)

### Batch Operations
- **batch_resize** - Resize multiple images concurrently
- **batch_compress** - Compress multiple images
- **batch_convert** - Convert multiple images to new format

### Utilities
- **list_models** - List available AI models with details

## Usage Examples

Once installed and configured, you can use natural language in Claude:

**Image Generation**:
- "Generate an image of a sunset over mountains"
- "Create a photorealistic portrait of a wise old wizard"
- "Generate an anime-style image of cherry blossoms"

**Image Processing**:
- "Resize photo.jpg to 800x600"
- "Compress all images in the photos directory to 85% quality"
- "Convert image.png to WebP format"

**Batch Operations**:
- "Resize all images in the vacation-photos folder to 1920x1080"
- "Compress all PNG files to reduce storage space"
- "Convert all photos to WebP for web use"

## Requirements

- Node.js 18.0.0 or higher
- Gemini API key (free tier available) or Vertex AI access

## Supported Platforms

- macOS (Intel & Apple Silicon)
- Linux (x86_64 & ARM64)
- Windows (x86_64)

## Environment Variables

The server respects these environment variables:

- `GEMINI_API_KEY` - Gemini API key for image generation
- `VERTEX_API_KEY` - Vertex AI API key (Express Mode)
- `VERTEX_PROJECT` - GCP project ID for Vertex AI
- `VERTEX_LOCATION` - Vertex AI location (default: us-central1)
- `GOOGLE_APPLICATION_CREDENTIALS` - Path to service account JSON

## Troubleshooting

### Server not connecting

1. Verify gimage is installed:
   ```bash
   which gimage
   ```

2. Test manually:
   ```bash
   gimage serve
   ```

3. Check Claude Desktop logs

### Image generation fails

1. Verify API key is configured:
   ```bash
   gimage auth gemini
   ```

2. Test generation manually:
   ```bash
   gimage generate "test image"
   ```

## Alternative Installation Methods

### Homebrew (macOS/Linux)

```bash
brew install chadneal/tap/gimage
```

Then configure Claude Desktop with:

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

### Manual Binary Download

Download from: https://github.com/chadneal/gimage/releases

## Documentation

- [Complete MCP Tools Reference](https://github.com/chadneal/gimage/blob/main/docs/MCP_TOOLS.md)
- [Usage Guide](https://github.com/chadneal/gimage/blob/main/docs/MCP_USAGE.md)
- [Examples](https://github.com/chadneal/gimage/blob/main/docs/MCP_EXAMPLES.md)
- [Main Documentation](https://github.com/chadneal/gimage)

## License

MIT

## Support

- GitHub Issues: https://github.com/chadneal/gimage/issues
- Documentation: https://github.com/chadneal/gimage#readme
