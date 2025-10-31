# Gimage Command Reference

Complete reference for all gimage commands, flags, and options.

## Table of Contents

- [Global Flags](#global-flags)
- [generate](#generate) - Generate images from text using AI
- [resize](#resize) - Resize images to dimensions
- [scale](#scale) - Scale images by factor
- [crop](#crop) - Crop images to regions
- [compress](#compress) - Compress images
- [convert](#convert) - Convert image formats
- [batch](#batch) - Batch process images
- [auth](#auth) - Configure authentication
  - [auth gemini](#auth-gemini)
  - [auth vertex](#auth-vertex)
- [config](#config) - Manage configuration
- [completion](#completion) - Generate shell completions

---

## Global Flags

These flags work with any command:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `$HOME/.gimage/config.md` |
| `--verbose` | Enable verbose output | `false` |
| `-h, --help` | Show help for command | - |
| `-v, --version` | Show version | - |

**Examples:**
```bash
gimage generate "sunset" --verbose
gimage --config ~/custom-config.md generate "landscape"
gimage --version
```

---

## generate

Generate images from text prompts using Google Gemini or Vertex AI.

### Usage
```bash
gimage generate [prompt] [flags]
```

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--api` | string | API to use: `gemini` or `vertex` | Auto-detected from model |
| `--api-key` | string | Gemini API key | From env/config |
| `--model` | string | Model to use | `gemini-2.5-flash-image` |
| `--size` | string | Image size (WxH) | `1024x1024` |
| `--style` | string | Style: `photorealistic`, `artistic`, `anime` | - |
| `--negative` | string | Negative prompt to avoid features | - |
| `--seed` | int | Random seed for reproducibility | `0` (random) |
| `-o, --output` | string | Output file path | `generated_<timestamp>.png` |
| `--project` | string | Vertex AI project ID | From env/config |
| `--location` | string | Vertex AI location | `us-central1` |
| `--list-models` | bool | List all available models and exit | `false` |

### Available Models

**Gemini API:**
- `gemini-2.5-flash-image` (default, recommended)
- `gemini-2.0-flash-preview-image-generation`

**Vertex AI:**
- `imagen-3.0-generate-002`
- `imagen-4`

### Examples

**Basic generation:**
```bash
gimage generate "a sunset over mountains"
```

**With specific size and style:**
```bash
gimage generate "futuristic city" --size 1024x1024 --style photorealistic
```

**Using Vertex AI:**
```bash
gimage generate "abstract art" --api vertex --model imagen-4
```

**With negative prompts:**
```bash
gimage generate "forest scene" --negative "people, buildings, cars"
```

**Reproducible generation:**
```bash
gimage generate "random pattern" --seed 12345
```

**Custom output path:**
```bash
gimage generate "landscape" --output my-landscape.png
```

**List all models:**
```bash
gimage generate --list-models
```

### Image Sizes

Common sizes (format: `WIDTHxHEIGHT`):
- `512x512` - Small, fast generation
- `1024x1024` - Default, balanced quality/speed
- `1536x1536` - High quality
- `2048x2048` - Maximum quality (Vertex AI only)

Custom sizes are supported by some models.

---

## resize

Resize images to specific dimensions using high-quality Lanczos resampling.

### Usage
```bash
gimage resize [input] [width] [height] [flags]
```

### Arguments

| Argument | Type | Description | Required |
|----------|------|-------------|----------|
| `input` | string | Input image file path | Yes |
| `width` | int | Target width in pixels | Yes |
| `height` | int | Target height in pixels | Yes |

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `-o, --output` | string | Output file path | `<input>_resized.<ext>` |

### Examples

**Basic resize:**
```bash
gimage resize photo.jpg 800 600
```

**Resize with custom output:**
```bash
gimage resize photo.png 1920 1080 --output fullhd.png
```

**Resize to thumbnail:**
```bash
gimage resize image.jpg 150 150 --output thumbnail.jpg
```

### Notes
- Maintains aspect ratio if only one dimension specified
- Uses Lanczos resampling for highest quality
- Preserves transparency for PNG images
- Output format matches input unless specified

---

## scale

Scale images by a factor (e.g., 0.5 for half size, 2.0 for double).

### Usage
```bash
gimage scale [input] [factor] [flags]
```

### Arguments

| Argument | Type | Description | Required |
|----------|------|-------------|----------|
| `input` | string | Input image file path | Yes |
| `factor` | float | Scale factor (e.g., 0.5, 2.0) | Yes |

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `-o, --output` | string | Output file path | `<input>_scaled.<ext>` |

### Examples

**Scale to 50% size:**
```bash
gimage scale photo.jpg 0.5
```

**Double the size:**
```bash
gimage scale image.png 2.0
```

**Scale with custom output:**
```bash
gimage scale photo.jpg 1.5 --output larger.jpg
```

### Notes
- Factor < 1.0 reduces size
- Factor > 1.0 increases size
- Factor = 1.0 creates a copy
- Uses Lanczos resampling for quality

---

## crop

Crop images to a specific region defined by x, y coordinates and dimensions.

### Usage
```bash
gimage crop [input] [x] [y] [width] [height] [flags]
```

### Arguments

| Argument | Type | Description | Required |
|----------|------|-------------|----------|
| `input` | string | Input image file path | Yes |
| `x` | int | Starting X coordinate | Yes |
| `y` | int | Starting Y coordinate | Yes |
| `width` | int | Crop width in pixels | Yes |
| `height` | int | Crop height in pixels | Yes |

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `-o, --output` | string | Output file path | `<input>_cropped.<ext>` |

### Examples

**Basic crop:**
```bash
gimage crop photo.jpg 100 100 800 600
```

**Crop from top-left corner:**
```bash
gimage crop image.png 0 0 1920 1080
```

**Crop with custom output:**
```bash
gimage crop photo.jpg 50 50 500 500 --output square.jpg
```

### Notes
- Coordinates start at (0, 0) in top-left corner
- Region must be within image bounds
- Preserves format and transparency

---

## compress

Compress images to reduce file size while maintaining quality.

### Usage
```bash
gimage compress [input] [flags]
```

### Arguments

| Argument | Type | Description | Required |
|----------|------|-------------|----------|
| `input` | string | Input image file path | Yes |

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--quality` | int | Compression quality (1-100) | `90` |
| `-o, --output` | string | Output file path | `<input>_compressed.<ext>` |

### Examples

**Basic compression:**
```bash
gimage compress photo.jpg
```

**Custom quality:**
```bash
gimage compress photo.jpg --quality 85
```

**Compress with output:**
```bash
gimage compress large.png --quality 75 --output small.png
```

### Quality Guidelines

| Quality | Use Case | File Size |
|---------|----------|-----------|
| 95-100 | Professional photography | Largest |
| 85-94 | High-quality web images | Large |
| 75-84 | Standard web images | Medium |
| 60-74 | Thumbnails, previews | Small |
| 1-59 | Maximum compression | Smallest |

### Notes
- Quality 90 is recommended default
- JPEG compression is lossy
- PNG uses lossless compression
- Higher quality = larger file size

---

## convert

Convert images to different formats.

### Usage
```bash
gimage convert [input] [format] [flags]
```

### Arguments

| Argument | Type | Description | Required |
|----------|------|-------------|----------|
| `input` | string | Input image file path | Yes |
| `format` | string | Target format | Yes |

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `-o, --output` | string | Output file path | `<input>.<format>` |

### Supported Formats

| Format | Extension | Description |
|--------|-----------|-------------|
| PNG | `.png` | Lossless, supports transparency |
| JPEG/JPG | `.jpg`, `.jpeg` | Lossy, best for photos |
| WebP | `.webp` | Modern, efficient |
| GIF | `.gif` | Animated images |
| TIFF | `.tiff`, `.tif` | Professional/archival |
| BMP | `.bmp` | Uncompressed |

### Examples

**PNG to JPEG:**
```bash
gimage convert image.png jpg
```

**JPEG to WebP:**
```bash
gimage convert photo.jpg webp
```

**With custom output:**
```bash
gimage convert input.png webp --output optimized.webp
```

### Notes
- Transparency is preserved when converting to PNG
- Converting to JPEG from PNG with transparency adds white background
- WebP offers best compression for web use
- Format is case-insensitive

---

## batch

Batch process multiple images with concurrent operations.

### Usage
```bash
gimage batch [operation] [input-dir] [flags]
```

### Arguments

| Argument | Type | Description | Required |
|----------|------|-------------|----------|
| `operation` | string | Operation: `resize`, `scale`, `crop`, `compress`, `convert` | Yes |
| `input-dir` | string | Input directory path | Yes |

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `-o, --output` | string | Output directory path | Same as input |
| `--workers` | int | Number of parallel workers | `4` |
| `--width` | int | Width for resize operation | - |
| `--height` | int | Height for resize operation | - |
| `--quality` | int | Quality for compress operation | `90` |

### Examples

**Resize all images:**
```bash
gimage batch resize photos/ --width 800 --height 600
```

**Resize with output directory:**
```bash
gimage batch resize photos/ --width 1920 --height 1080 --output resized/
```

**Compress all images:**
```bash
gimage batch compress photos/ --quality 85 --output compressed/
```

**Convert all to WebP:**
```bash
gimage batch convert photos/ webp --output webp/
```

**Use 8 workers for faster processing:**
```bash
gimage batch resize photos/ --width 800 --height 600 --workers 8
```

### Supported Operations

| Operation | Description | Required Flags |
|-----------|-------------|----------------|
| `resize` | Resize to dimensions | `--width`, `--height` |
| `scale` | Scale by factor | (factor as argument) |
| `compress` | Compress images | `--quality` (optional) |
| `convert` | Convert format | (format as argument) |

### Notes
- Processes all supported image formats in directory
- Default 4 workers (adjust based on CPU cores)
- Creates output directory if it doesn't exist
- Skips unsupported files automatically
- Shows progress for each file
- Preserves directory structure in output

---

## auth

Manage authentication for Gemini and Vertex AI.

### Usage
```bash
gimage auth [subcommand]
```

### Subcommands
- `gemini` - Configure Gemini API authentication
- `vertex` - Configure Vertex AI authentication

---

### auth gemini

Interactive setup for Gemini API authentication.

### Usage
```bash
gimage auth gemini
```

### Interactive Prompts
1. **Gemini API Key** - Your API key from AI Studio

### Process
1. Loads existing config values as defaults
2. Prompts for API key (shows masked preview if existing)
3. Saves to `~/.gimage/config.md` with 0600 permissions
4. Confirms successful configuration

### Get API Key
Visit: https://aistudio.google.com/app/apikey

### Example Session
```bash
$ gimage auth gemini
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Gemini API Authentication Setup
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Get your API key from: https://aistudio.google.com/app/apikey

Gemini API Key [...cVVI]: <paste your key or press Enter>

✓ Configuration saved successfully!
  Location: /Users/you/.gimage/config.md

You can now use Gemini API with:
  gimage generate "your prompt here"
```

---

### auth vertex

Interactive setup for Vertex AI authentication.

### Usage
```bash
gimage auth vertex
```

### Authentication Modes

**Mode 1: Express Mode (API Key)**
- Simplest setup
- Good for development and testing
- Get API key from Google Cloud Console

**Mode 2: Full Mode (Service Account)**
- Production-ready
- Fine-grained access control
- Requires service account JSON file

**Mode 3: Full Mode (ADC)**
- Local development
- Uses your gcloud credentials
- Requires `gcloud auth application-default login`

### Interactive Prompts

**All Modes:**
1. Choose authentication mode (1, 2, or 3)

**Mode 1 (Express):**
1. Vertex AI API Key
2. Google Cloud Project ID (optional)
3. Location/Region

**Mode 2 (Service Account):**
1. Google Cloud Project ID
2. Location/Region
3. Path to service account JSON file

**Mode 3 (ADC):**
1. Google Cloud Project ID
2. Location/Region

### Example Session (Express Mode)
```bash
$ gimage auth vertex
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Vertex AI Authentication Setup
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Choose your authentication mode:

  1. Express Mode - API Key (simple, good for testing)
     • Sign up at: https://console.cloud.google.com/vertex-ai
     • Get API key from: APIs & Services > Credentials
     • Best for: Development, testing, rapid prototyping

  2. Full Mode - Service Account (secure, production-ready)
     • Requires: GCP project, service account JSON file
     • Best for: Production, fine-grained access control

  3. Full Mode - Application Default Credentials (local dev)
     • Run: gcloud auth application-default login
     • Best for: Local development with your GCP account

Choose mode (1, 2, or 3) [1]: 1

━━━ Express Mode Setup ━━━

Get your API key:
  1. Go to: https://console.cloud.google.com/vertex-ai
  2. Sign up for Vertex AI Express Mode
  3. Find your API key in: APIs & Services > Credentials

Vertex AI API Key: <paste your key>
Google Cloud Project ID (optional): my-project
Location/Region [us-central1]:

✓ Express Mode configured successfully!
  Location: /Users/you/.gimage/config.md

You can now use Vertex AI with:
  gimage generate --api vertex "your prompt here"
```

### Get Started with Vertex AI

**Express Mode:**
1. Visit: https://console.cloud.google.com/vertex-ai
2. Sign up for Vertex AI Express Mode
3. Get API key from: APIs & Services > Credentials

**Service Account:**
1. Create GCP project
2. Enable Vertex AI API
3. Create service account
4. Download JSON key file

**ADC:**
```bash
gcloud auth application-default login
```

---

## config

Manage gimage configuration.

### Usage
```bash
gimage config [flags]
```

### Description
Currently a placeholder for future configuration management commands.
Use `gimage auth` commands to set up authentication.

### Configuration File

**Location:** `~/.gimage/config.md`

**Format:**
```markdown
# Gimage Configuration

**gemini_api_key**: AIzaSy...
**vertex_api_key**: AIzaSy...
**vertex_project**: your-project-id
**vertex_location**: us-central1
**vertex_credentials_path**: ~/.gimage/credentials/service-account.json
**default_api**: gemini
**default_model**: gemini-2.5-flash-image
**default_size**: 1024x1024
**log_level**: info
```

### Manual Editing
You can manually edit `~/.gimage/config.md` if needed:
```bash
# Edit with your preferred editor
nano ~/.gimage/config.md
vim ~/.gimage/config.md
code ~/.gimage/config.md
```

Format: `**key**: value` on each line

---

## completion

Generate shell completion scripts for bash, zsh, fish, or PowerShell.

### Usage
```bash
gimage completion [shell]
```

### Supported Shells
- `bash`
- `zsh`
- `fish`
- `powershell`

### Examples

**Bash:**
```bash
gimage completion bash > /etc/bash_completion.d/gimage
```

**Zsh:**
```bash
gimage completion zsh > "${fpath[1]}/_gimage"
```

**Fish:**
```bash
gimage completion fish > ~/.config/fish/completions/gimage.fish
```

**PowerShell:**
```powershell
gimage completion powershell | Out-String | Invoke-Expression
```

### Setup Instructions

**Bash (Linux):**
```bash
gimage completion bash | sudo tee /etc/bash_completion.d/gimage
source /etc/bash_completion.d/gimage
```

**Bash (macOS):**
```bash
brew install bash-completion
gimage completion bash > $(brew --prefix)/etc/bash_completion.d/gimage
```

**Zsh:**
```bash
gimage completion zsh > ~/.zsh/completions/_gimage
# Add to ~/.zshrc:
fpath=(~/.zsh/completions $fpath)
autoload -U compinit && compinit
```

**Fish:**
```bash
gimage completion fish > ~/.config/fish/completions/gimage.fish
```

---

## Environment Variables

These environment variables can be used to configure gimage:

| Variable | Description | Example |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Gemini API key | `AIzaSy...` |
| `VERTEX_API_KEY` | Vertex AI Express Mode key | `AIzaSy...` |
| `VERTEX_PROJECT` | Vertex AI project ID | `my-gcp-project` |
| `VERTEX_LOCATION` | Vertex AI location | `us-central1` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Service account file path | `/path/to/key.json` |
| `GIMAGE_CONFIG` | Custom config file path | `~/my-config.md` |
| `GIMAGE_LOG_LEVEL` | Log level (debug, info, warn, error) | `debug` |

### Priority Order
1. Command-line flags (highest)
2. Environment variables
3. Config file
4. Default values (lowest)

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Command-line usage error |
| 130 | Interrupted (Ctrl+C) |

---

## Tips & Tricks

### Reproducible Generations
Use `--seed` to get the same image every time:
```bash
gimage generate "abstract art" --seed 42
```

### Quick Thumbnails
Create thumbnails for all images:
```bash
gimage batch resize photos/ --width 200 --height 200 --output thumbnails/
```

### Optimize for Web
Compress and convert to WebP:
```bash
gimage batch convert photos/ webp --output web/
gimage batch compress web/ --quality 85
```

### Test Different Models
Compare model outputs:
```bash
gimage generate "sunset" --model gemini-2.5-flash-image --output gemini.png
gimage generate "sunset" --api vertex --model imagen-4 --output imagen.png
```

### Process Only JPEGs
Use shell globbing:
```bash
for file in *.jpg; do gimage compress "$file" --quality 80; done
```

### Chain Operations
Process images in pipeline:
```bash
gimage generate "landscape" --output temp.png
gimage resize temp.png 1920 1080 --output hd.png
gimage compress hd.png --quality 90 --output final.png
```

---

## Common Workflows

### Social Media Content
```bash
# Generate image
gimage generate "tech background" --size 1024x1024

# Create Instagram post (1080x1080)
gimage resize generated.png 1080 1080 --output instagram.png

# Create Twitter header (1500x500)
gimage crop generated.png 0 262 1500 500 --output twitter-header.png
```

### Batch Optimization
```bash
# Resize all photos to HD
gimage batch resize photos/ --width 1920 --height 1080 --output hd/

# Compress for web
gimage batch compress hd/ --quality 85 --output web/

# Convert to WebP
gimage batch convert web/ webp --output optimized/
```

### AI Art Generation
```bash
# Generate multiple variations
for i in {1..5}; do
  gimage generate "abstract landscape" --seed $i --output "art-$i.png"
done

# Generate different styles
gimage generate "cityscape" --style photorealistic --output photo.png
gimage generate "cityscape" --style artistic --output artistic.png
gimage generate "cityscape" --style anime --output anime.png
```

---

## Troubleshooting

### "API key not configured"
Run authentication setup:
```bash
gimage auth gemini
# or
gimage auth vertex
```

### "Failed to load config"
Check config file permissions:
```bash
ls -la ~/.gimage/config.md
chmod 600 ~/.gimage/config.md
```

### "Model not found"
List available models:
```bash
gimage generate --list-models
```

### "Image too large"
Reduce size or use scaling:
```bash
gimage scale large-image.jpg 0.5 --output smaller.jpg
```

### Verbose Mode
Enable detailed logging:
```bash
gimage generate "test" --verbose
```

---

## More Information

- **GitHub**: https://github.com/chadneal/gimage
- **Issues**: https://github.com/chadneal/gimage/issues
- **Discussions**: https://github.com/chadneal/gimage/discussions
- **Main Documentation**: See [README.md](README.md)
- **Developer Guide**: See [CLAUDE.md](CLAUDE.md)
