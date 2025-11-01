# Using Gimage with AI Assistants

Complete guide to using the gimage MCP server with Claude and other AI assistants.

## Quick Start

### 1. Install gimage MCP server

**Option A: npm (Recommended)**
```bash
npm install -g @apresai/gimage-mcp
```

**Option B: Homebrew**
```bash
brew install apresai/tap/gimage
```

**Option C: Manual Download**

Download from [GitHub Releases](https://github.com/apresai/gimage/releases)

### 2. Configure API credentials

Before using the MCP server, you need to set up authentication for AI image generation:

```bash
# Setup Gemini API (recommended - free tier available)
gimage auth gemini
```

**Get your free API key**: https://aistudio.google.com/app/apikey

The Gemini free tier includes:
- 1500 requests per day
- All Gemini models
- No credit card required

**Alternative: Vertex AI** (for advanced users)
```bash
gimage auth vertex
```

### 3. Add to Claude Desktop

Edit your Claude Desktop MCP configuration file:

**Configuration File Locations:**
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**Add this configuration:**

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

**If using Homebrew installation:**

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

### 4. Restart Claude Desktop

Quit Claude Desktop completely and reopen it. The MCP server will start automatically.

### 5. Start Using!

Try these prompts in Claude:

**Image Generation:**
- "Generate an image of a sunset over mountains"
- "Create a photorealistic portrait of a wise old wizard"
- "Generate an anime-style image of cherry blossoms in spring"

**Image Processing:**
- "Resize photo.jpg to 800x600 pixels"
- "Compress all images in my photos folder to save space"
- "Convert image.png to WebP format for web use"

---

## Available Operations

### 🎨 AI Image Generation

Generate stunning images from text descriptions using state-of-the-art AI models.

**Capabilities:**
- **Multiple AI models**: Gemini 2.5 Flash (default), Gemini 2.0 Flash, Imagen 3, Imagen 4
- **Size options**: From 256x256 up to 2048x2048 pixels
- **Style controls**: Photorealistic, artistic, anime
- **Negative prompts**: Exclude unwanted elements
- **Reproducible results**: Use seeds for consistent generation

**Example Prompts:**
```
"Generate a 1024x1024 photorealistic image of a medieval castle on a hill"
"Create an artistic interpretation of a futuristic city at night"
"Generate an anime-style character with blue hair and a sword"
"Make me an image of a sunset, but avoid showing any people or buildings"
```

### 🖼️ Image Processing

#### Resize
Change image dimensions to exact pixel values.

**Example Prompts:**
```
"Resize landscape.jpg to 1920x1080"
"Make photo.png exactly 800 pixels wide and 600 pixels tall"
```

#### Scale
Resize image proportionally by a factor.

**Example Prompts:**
```
"Scale photo.jpg to half its size"
"Make image.png twice as large"
"Reduce all dimensions by 75% (use factor 0.25)"
```

#### Crop
Extract a specific region from an image.

**Example Prompts:**
```
"Crop photo.jpg starting at position (100, 100) with width 800 and height 600"
"Extract a 500x500 square from the center of the image"
```

#### Compress
Reduce file size while maintaining quality.

**Example Prompts:**
```
"Compress photo.jpg to 85% quality to save space"
"Reduce the file size of all images in this directory"
```

#### Convert
Change image format (PNG, JPG, WebP, GIF, TIFF, BMP).

**Example Prompts:**
```
"Convert photo.png to JPEG format"
"Change all images to WebP for better web performance"
```

### ⚡ Batch Operations

Process multiple images concurrently for efficient workflows.

**Example Prompts:**
```
"Resize all images in the vacation-photos folder to 1920x1080"
"Compress every image in my-photos directory to 85% quality"
"Convert all PNG files in this directory to WebP"
```

**Features:**
- Concurrent processing with multiple CPU cores
- Progress reporting
- Error handling (continues even if some files fail)
- Preserves directory structure

---

## Common Workflows

### Web Image Optimization

**Scenario**: You have a folder of photos that need to be optimized for web use.

**Prompt**:
```
"I have photos in the 'website-images' folder. Please:
1. Resize them all to a maximum of 1920x1080
2. Compress them to 85% quality
3. Convert them to WebP format
Save the results in 'optimized-images' folder"
```

Claude will execute these operations in sequence.

### Social Media Preparation

**Scenario**: Prepare an image for Instagram.

**Prompt**:
```
"Take photo.jpg and prepare it for Instagram:
- Crop it to a square (1080x1080) from the center
- Compress to 90% quality
Save it as instagram.jpg"
```

### E-commerce Product Images

**Scenario**: Generate multiple sizes of a product image.

**Prompt**:
```
"I have product.jpg. Create three versions:
1. Large: 1200x1200 (save as product-large.jpg)
2. Medium: 600x600 (save as product-medium.jpg)
3. Thumbnail: 200x200 (save as product-thumb.jpg)
All should be compressed to 90% quality"
```

### AI Content Generation

**Scenario**: Generate multiple variations of an image.

**Prompt**:
```
"Generate 3 different versions of a fantasy landscape:
1. A photorealistic mountain scene with a lake
2. An artistic interpretation with vibrant colors
3. An anime-style version with dramatic clouds
Use 1024x1024 size for all"
```

---

## Troubleshooting

### MCP Server Not Connecting

**Symptoms**: Claude shows no gimage tools available

**Solutions**:

1. **Verify gimage is installed**
   ```bash
   which gimage
   ```
   Should show the path to gimage binary.

2. **Test manual startup**
   ```bash
   gimage serve
   ```
   Should start without errors. Press Ctrl+C to stop.

3. **Check Claude Desktop configuration**
   - Verify JSON syntax is correct
   - Ensure file path matches your system
   - Check for typos in command/args

4. **Restart Claude Desktop**
   - Quit completely (not just close window)
   - Reopen application

5. **Check Claude logs**
   - macOS: `~/Library/Logs/Claude/`
   - Look for error messages related to gimage

### Image Generation Fails

**Symptoms**: Error messages when trying to generate images

**Solutions**:

1. **Verify API key is configured**
   ```bash
   gimage auth gemini
   ```
   Follow the prompts to enter your API key.

2. **Test generation manually**
   ```bash
   gimage generate "test image"
   ```
   If this works, MCP server should work too.

3. **Check API key validity**
   - Ensure key hasn't expired
   - Verify it's correctly copied (no extra spaces)
   - Get a new key if needed: https://aistudio.google.com/app/apikey

4. **Verify internet connection**
   - Gemini API requires internet access
   - Check firewall settings

### Permission Errors

**Symptoms**: "Permission denied" when processing images

**Solutions**:

1. **Check file permissions**
   ```bash
   ls -la /path/to/image.jpg
   ```

2. **Ensure write access to output directory**
   ```bash
   mkdir -p output-folder
   chmod 755 output-folder
   ```

3. **Use absolute paths**
   Instead of: `photo.jpg`
   Use: `/Users/yourname/photos/photo.jpg`

### Batch Operations Slow

**Symptoms**: Batch processing takes too long

**Solutions**:

1. **Increase workers**
   "Process all images in photos/ using 8 workers for faster processing"

2. **Process in smaller batches**
   Split large directories into smaller chunks

3. **Check disk space**
   Ensure sufficient space for output files

---

## Advanced Usage

### Environment Variables

You can override configuration with environment variables:

```bash
# Gemini API
export GEMINI_API_KEY="your-key-here"

# Vertex AI Express Mode
export VERTEX_API_KEY="your-vertex-key"
export VERTEX_PROJECT="your-gcp-project"
export VERTEX_LOCATION="us-central1"

# Vertex AI Full Mode (Service Account)
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
export VERTEX_PROJECT="your-gcp-project"
```

### Using Different AI Models

**Gemini Models** (Fast, Free Tier):
- `gemini-2.5-flash-image` (default, recommended)
- `gemini-2.0-flash-preview-image-generation`

**Vertex AI Models** (Premium Quality):
- `imagen-3.0-generate-002`
- `imagen-4` (highest quality, up to 2K resolution)

**Example**:
```
"Generate an image using Imagen 4 model with 2048x2048 resolution showing a hyper-realistic dragon"
```

### Reproducible Generation

Use seeds for consistent results:

```
"Generate an image with seed 12345 of a mountain landscape"
"Generate the same image again with seed 12345 to verify it's identical"
```

### Complex Multi-Step Workflows

Claude can chain multiple operations:

```
"Please help me with this workflow:
1. Generate an AI image of a fantasy castle (1024x1024)
2. Create 3 different sizes: 2048x2048, 1024x1024, 512x512
3. Compress all of them to 90% quality
4. Convert to WebP format
5. Save in a folder called 'castle-variants'"
```

---

## Tips for Best Results

### Image Generation

1. **Be specific** - Detailed prompts produce better results
   - ✅ "A photorealistic sunset over mountain peaks with orange and purple sky, reflection in a calm lake"
   - ❌ "Sunset"

2. **Use style keywords** - Specify the artistic style
   - "photorealistic", "artistic", "anime", "painting", "sketch", etc.

3. **Use negative prompts** - Exclude unwanted elements
   - "Generate a forest scene, but avoid showing any people, buildings, or modern objects"

4. **Try different models** - Each has strengths
   - Gemini 2.5 Flash: Fast, great for most uses
   - Imagen 4: Highest quality, best for professional work

### Image Processing

1. **Know your target** - Understand final use case
   - Web: 1920x1080 or smaller, WebP format, 85% quality
   - Print: Larger sizes, PNG/TIFF, 95-100% quality
   - Mobile: 750x1334, 80% quality

2. **Batch similar operations** - More efficient
   - Process all images in a folder at once
   - Use consistent settings across related images

3. **Preserve originals** - Always work on copies
   - Use explicit output paths
   - Create separate directories for processed images

4. **Check results** - Verify quality meets needs
   - View processed images before deleting originals
   - Adjust quality settings if needed

---

## Next Steps

- **Explore all tools**: Try each operation to understand capabilities
- **Create workflows**: Combine operations for your specific needs
- **Share examples**: Help others by sharing successful prompts
- **Report issues**: Help improve the tool by reporting bugs

---

## Support

- **Documentation**: [Complete Tools Reference](MCP_TOOLS.md)
- **Examples**: [Real-World Examples](MCP_EXAMPLES.md)
- **GitHub**: https://github.com/apresai/gimage
- **Issues**: https://github.com/apresai/gimage/issues
