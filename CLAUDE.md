# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

`gimage` - A Go-based CLI tool for AI-powered image generation and processing.

**Core Capabilities**:
- Generate images using Google Gemini 2.5 Flash, Vertex AI Imagen 4, or AWS Bedrock Nova Canvas
- Process images: resize, scale, crop, compress, convert (PNG, JPG, WebP, GIF, TIFF, BMP)
- Batch processing with concurrent operations
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
- `generate_image`, `resize_image`, `scale_image`, `crop_image`, `compress_image`
- `convert_image`, `batch_resize`, `batch_compress`, `batch_convert`, `list_models`

Config: `~/.gimage/config.md` (markdown format using `**key**: value`)

## Authentication

Interactive commands create/update config:

```bash
gimage auth gemini    # Gemini API key
gimage auth vertex    # Vertex AI (3 modes: Express, Service Account, ADC)
gimage auth bedrock   # AWS Bedrock (2 modes: REST with keys, SDK with credential chain)
```

Config file format (`~/.gimage/config.md`):
```markdown
**gemini_api_key**: AIzaSy...
**vertex_api_key**: AIzaSy...
**vertex_project**: your-project-id
**vertex_location**: us-central1
**aws_access_key_id**: AKIA...
**aws_secret_access_key**: wJalr...
**aws_region**: us-east-1
**default_api**: gemini
**default_model**: gemini-2.5-flash-image
```

## Security & Best Practices

### Documentation and Dates
- **ALWAYS use `date +%Y-%m-%d`** command for current date
- Never hardcode dates in documentation
- Use dynamic date retrieval for CHANGELOG.md and docs

### Credentials
- Never log API keys
- Config file created with 0600 permissions
- Use `gimage auth` commands instead of manual editing
- Use environment variables for CI/CD

### Code Quality
- Follow Go idioms and conventions
- Keep functions small and focused
- Use golangci-lint
- Document all public APIs with godoc

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
