# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

`gimage` - A Go-based CLI tool for AI-powered image generation and processing.

**Core Capabilities**:
- Generate images using Google Gemini 2.5 Flash, Vertex AI Imagen 4, or AWS Bedrock Nova Canvas
- Process images: resize, scale, crop, compress, convert (PNG, JPG, WebP, GIF, TIFF, BMP)
- Batch processing via MCP server (batch_resize, batch_compress, batch_convert)
- MCP server for Claude Desktop integration
- AWS Lambda API deployment

**Technology Stack**:
- Pure Go 1.22+ (zero C dependencies for portability)
- Image processing: `github.com/disintegration/imaging`
- CLI: Cobra + Viper
- APIs: Gemini API, Vertex AI, AWS Bedrock

## Build Commands

```bash
make build          # Build CLI binary
make build-all      # Build for all platforms
make install        # Install locally
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make clean          # Clean artifacts
make benchmark      # Run benchmarks
```

## Project Structure

```
gimage/
├── cmd/gimage/              # CLI entrypoint
├── internal/
│   ├── imaging/             # Image processing operations
│   ├── generate/            # AI image generation (Gemini, Vertex, Bedrock)
│   ├── config/              # Configuration & authentication
│   ├── cli/                 # CLI commands
│   └── mcp/                 # MCP server implementation
├── pkg/models/              # Shared types
├── test/
│   ├── fixtures/            # Test images (DO NOT MODIFY)
│   └── integration/         # Integration tests
└── docs/                    # Documentation
```

## Architecture Patterns

### Pure Go Philosophy
This project uses **pure Go with zero C dependencies**:
- Single binary distribution, no system dependencies
- Cross-compilation to any platform
- Uses `disintegration/imaging` (not bimg/libvips)
- **Never add C library dependencies**

### Configuration Hierarchy (Priority Order)
1. Command-line flags (highest)
2. Environment variables (`GEMINI_API_KEY`, `VERTEX_API_KEY`, `AWS_ACCESS_KEY_ID`, etc.)
3. Config file (`~/.gimage/config.md`)
4. Default values (lowest)

### API Client Pattern
All backends (Gemini, Vertex, Bedrock) implement common interface:
```go
type ImageGenerator interface {
    GenerateImage(ctx context.Context, prompt string, options GenerateOptions) (*GeneratedImage, error)
    Close() error
}
```

### Error Handling
- Return errors with context using `fmt.Errorf` with `%w`
- Provide actionable error messages
- Never panic in production code
- Validate inputs early

## Multi-Backend Architecture

**Supported Backends**:
- **Gemini API** (REST) - Free tier, fastest setup
- **Vertex AI** - Express Mode (REST) or Full Mode (SDK)
- **AWS Bedrock** - REST or SDK modes

### Backend Selection Logic

Model name implies backend (auto-detect):
- `gemini-2.5-flash-image` → gemini
- `imagen-4` → vertex
- `amazon.nova-canvas-v1:0` → bedrock

Optional `--api` flag overrides auto-detection.

### Model Name Resolution

Map informal names to exact model IDs:

| User Input | Exact Model ID | API |
|-----------|---------------|-----|
| "gemini", "flash" | `gemini-2.5-flash-image` | gemini |
| "imagen", "imagen-4" | `imagen-4` | vertex |
| "nova", "nova-canvas" | `amazon.nova-canvas-v1:0` | bedrock |

**Always use exact model IDs from the mapping table.**

## Development Workflow

### Adding a New CLI Command
1. Create command file in `internal/cli/`
2. Implement using Cobra patterns
3. Add flags with Viper binding
4. Wire up to root command
5. Add unit tests
6. Update `COMMANDS.md`

### Adding Image Processing Operations
1. Create operation file in `internal/imaging/`
2. Use `disintegration/imaging` library exclusively
3. Handle all supported formats (PNG, JPG, WebP, GIF, TIFF, BMP)
4. Add comprehensive error handling
5. Create unit tests with fixtures from `test/fixtures/` (DO NOT MODIFY)
6. Benchmark critical operations

## CLI Standards

### Command Interface Pattern

**Image processing commands use explicit flags**:
- Consistent, explicit, self-documenting
- Composable with shell scripts
- Clear in logs and command history

**Generation command supports both positional and flag-based prompts**:
```bash
# Positional prompt (most common, recommended for quick use)
gimage generate "sunset over mountains"

# Flag-based prompt (explicit, useful in scripts)
gimage generate --prompt "sunset over mountains"
```

**Standard Flags**:
- `--input, -i`: Input file path (required for most image processing commands)
- `--output, -o`: Output file path (optional, auto-generated if omitted)
- `--verbose, -v`: Enable verbose output (available on all commands)

**Examples**:
```bash
# Image processing commands (flags-only)
gimage resize --input photo.jpg --width 800 --height 600 --output resized.jpg
gimage crop --input photo.jpg --x 100 --y 100 --width 400 --height 300
gimage scale --input photo.jpg --factor 0.5
gimage convert --input photo.jpg --format webp
gimage compress --input photo.jpg --quality 85

# Generation command (supports both styles)
gimage generate "sunset over mountains" --size 1024x1024
gimage generate --prompt "sunset over mountains" --output sunset.png

# Auth commands (positional provider argument)
gimage auth status
gimage auth setup gemini
gimage auth list
gimage auth test gemini
```

**Available CLI Commands**:
- `generate` - Generate images from text prompts
- `resize` - Resize images to specific dimensions
- `scale` - Scale images by a factor
- `crop` - Crop images to specific regions
- `compress` - Compress images to reduce file size (JPG, WebP)
- `convert` - Convert images between formats
- `auth` - Configure and manage API credentials
- `serve` - Start MCP server (includes batch operations)
- `tui` - Launch interactive terminal UI

**Removed Commands** (no longer available):
- `batch` - Use MCP server tools instead (batch_resize, batch_compress, batch_convert)
- `config` - Use `auth` commands for configuration

### Verbose Logging

All commands support `--verbose` flag for detailed output:
```bash
gimage resize --input photo.jpg --width 800 --height 600 --verbose
# Outputs:
# ℹ Resizing photo.jpg to 800x600...
# • Input: photo.jpg
# • Output: photo_resized_800x600.jpg
# • Dimensions: 800x600
# ✓ Resized successfully!
```

### Output Path Generation

If `--output` is omitted, commands auto-generate descriptive output paths:
- `resize`: `input_resized_WxH.ext`
- `crop`: `input_cropped_WxH.ext`
- `scale`: `input_scaled_FACTORx.ext`
- `convert`: `input_converted.FORMAT`
- `compress`: `input_compressed.ext`

### Testing Strategy

**Unit Tests (>80% coverage required)**:
- Test request building logic (validate JSON structure)
- Test response parsing with real example responses
- Test input validation (dimensions, prompts, parameters)
- Test configuration loading
- Test CLI flag parsing

**Integration Tests (manual, costs money)**:
- Real API calls to Gemini/Vertex/Bedrock
- Run manually: `go test -tags=integration`
- **DO NOT MOCK cloud provider APIs** - mocks provide zero value

**Table-driven tests** for multiple scenarios.

### MCP Server

MCP server runs via `gimage serve` and exposes 10 tools for AI assistants:
- **Single operations**: `generate_image`, `resize_image`, `scale_image`, `crop_image`, `compress_image`, `convert_image`
- **Batch operations**: `batch_resize`, `batch_compress`, `batch_convert` (concurrent processing)
- **Utilities**: `list_models`

**Important**: Batch operations are ONLY available through MCP server, not CLI.
- CLI users should wrap `gimage` in shell scripts for batch processing
- MCP server provides optimized concurrent batch operations for AI assistants

Config: `~/.gimage/config.md` (markdown format using `**key**: value`)

## Authentication

### Auth Commands

Modern auth command structure:

```bash
gimage auth status    # Show authentication status for all providers
gimage auth list      # List all configured providers with sources
gimage auth test      # Test credentials by making real API calls
gimage auth setup     # Interactive setup wizard for providers
```

### Authentication Precedence (Highest to Lowest)

**All Providers**:
1. Command-line flags (e.g., `--gemini-api-key`)
2. Environment variables
3. Config file (`~/.gimage/config.md`)
4. Default values

**Gemini API**:
- Single credential: `GEMINI_API_KEY`
- Simple REST client with API key

**Vertex AI** (3 authentication modes):
1. **Express Mode (REST)**: `VERTEX_API_KEY` → Fast, simple, REST-based
2. **Service Account**: `GOOGLE_APPLICATION_CREDENTIALS` → JSON key file path
3. **Application Default Credentials (ADC)**: Automatic → gcloud SDK, workload identity

**AWS Bedrock** (4 authentication modes):
1. **REST with Bearer Token**: `AWS_BEARER_TOKEN_BEDROCK` → Direct REST API
2. **SDK with Access Keys**: `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` → AWS SDK
3. **SDK with Profile**: `AWS_PROFILE` → Named profile from `~/.aws/credentials`
4. **SDK with IAM Role**: Automatic → EC2, Lambda, ECS task roles

**Why Both REST and SDK?**
- REST: Simple API keys, quick setup, perfect for local development
- SDK: IAM roles, profiles, workload identity - critical for Lambda/EC2/GCP deployments

### Config File Format

Location: `~/.gimage/config.md` (markdown format, 0600 permissions)

```markdown
# Gimage Configuration

⚠️  SECURITY WARNING ⚠️
This file contains SENSITIVE API KEYS stored in PLAINTEXT.

**gemini_api_key**: AIzaSy...
**vertex_api_key**: AIzaSy...
**vertex_project**: your-project-id
**vertex_location**: us-central1
**vertex_credentials_path**: /path/to/service-account.json
**aws_access_key_id**: AKIA...
**aws_secret_access_key**: wJalr...
**aws_region**: us-east-1
**aws_profile**: default
**aws_bedrock_api_key**: bearer-token-here
**default_api**: gemini
**default_model**: gemini-2.5-flash-image
**log_level**: info
```

## Security & Best Practices

### Credential Security

**Config File Security**:
- Config file (`~/.gimage/config.md`) stores API keys in **PLAINTEXT**
- File created with 0600 permissions (only owner can read/write)
- Includes prominent security warnings at the top
- **NEVER commit config file to version control**
- **NEVER share config file or its contents**

**Best Practices**:
- **PREFER environment variables** over config file for sensitive keys
- Use `gimage auth status` to see where credentials are coming from
- Rotate API keys regularly (every 90 days recommended)
- Use separate keys for dev/staging/production environments
- For CI/CD pipelines, always use environment variables
- For Lambda/EC2/ECS, prefer IAM roles over static credentials

**Environment Variable Priority**:
- Environment variables override config file (by design)
- Set `GEMINI_API_KEY`, `VERTEX_API_KEY`, `AWS_ACCESS_KEY_ID`, etc.
- Use `gimage auth status` to check for conflicts

**Warning About Conflicts**:
- If both config file AND environment variable are set, env var wins
- `gimage auth status` will warn you about conflicting credentials
- Clean up unused credentials to avoid confusion

### Documentation and Dates
- **ALWAYS use `date +%Y-%m-%d`** command for current date
- Never hardcode dates in documentation
- Use dynamic date retrieval for CHANGELOG.md and docs

### Code Quality
- Follow Go idioms and conventions
- Keep functions small and focused
- Use golangci-lint
- Document all public APIs with godoc
- Never log API keys or sensitive data

## Common Patterns

### Loading and Saving Images
```go
img, err := imaging.Open(inputPath)
if err != nil {
    return fmt.Errorf("failed to open image: %w", err)
}

result := imaging.Resize(img, width, height, imaging.Lanczos)
err = imaging.Save(result, outputPath)
```

### Concurrent Processing
```go
workers := runtime.NumCPU()
sem := make(chan struct{}, workers)
var wg sync.WaitGroup

for _, file := range files {
    wg.Add(1)
    go func(f string) {
        defer wg.Done()
        sem <- struct{}{}
        defer func() { <-sem }()
        // Process file
    }(file)
}
wg.Wait()
```

## Git Usage Policy

**IMPORTANT**: Do NOT use git commands unless user explicitly asks.

**Do NOT**:
- Auto-commit after creating/modifying files
- Auto-commit after completing features
- Auto-commit when you think "this should be committed"

**DO**:
- Only when user says "commit this"
- Only when user says "push to GitHub"
- Only when user explicitly requests git operations

**Why**: User controls when and how code is committed. Automatic commits interrupt workflow and create unwanted history.

## Release Process

1. Update version in code
2. Run `make test`
3. Build all platforms: `make build-all`
4. Create tag: `git tag v1.x.x` (only when user requests)
5. GitHub Actions handles release automation

## Lambda Deployment

Deploy as serverless REST API on AWS Lambda:

```bash
make build-lambda      # Build for ARM64/Graviton2
make package-lambda    # Create deployment zip
cd infrastructure/cdk && cdk deploy
```

See `lambda.md` for complete guide.

## Documentation Structure

- **README.md** - Main project overview
- **COMMANDS.md** - Full CLI command reference
- **lambda.md** - Lambda deployment guide
- **INTEGRATION_GUIDE.md** - API client examples
- **TESTING.md** - Testing documentation
- **mcp.md** - MCP server overview
- **docs/MCP_TOOLS.md** - Complete MCP tools reference (for LLMs)
- **docs/MCP_USAGE.md** - Primary MCP user guide (for LLMs)
- **docs/MCP_EXAMPLES.md** - Real-world MCP examples (for LLMs)

## Implementation Priorities

Core development phases:
1. Project initialization
2. Image processing core
3. AI API integrations (Gemini → Vertex → Bedrock)
4. CLI commands
5. Configuration system
6. Testing suite
7. Documentation
8. MCP server
9. Lambda deployment
10. Distribution (Homebrew, npm)
