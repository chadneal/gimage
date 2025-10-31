# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `gimage` - a Go-based CLI tool for AI-powered image generation and processing.

**Core Capabilities**:
- Generate images from text using Google Gemini 2.5 Flash Image or Vertex AI Imagen 4
- Process images: resize, scale, crop, compress (PNG, JPG, WebP, GIF, TIFF, BMP)
- Batch processing with concurrent operations
- MCP server for Claude integration

**Technology Stack**:
- Go 1.22+ (pure Go, zero C dependencies)
- Image processing: `github.com/disintegration/imaging`
- CLI framework: Cobra + Viper
- APIs: Gemini API and Vertex AI

## Build Commands

```bash
# Build the CLI
make build

# Build for all platforms
make build-all

# Install locally
make install

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Clean build artifacts
make clean

# Run benchmarks
make benchmark
```

## Project Structure

```
gimage/
├── cmd/gimage/              # CLI entrypoint
├── internal/
│   ├── imaging/             # Image processing (resize, scale, crop, compress)
│   ├── generate/            # AI image generation (Gemini & Vertex clients)
│   ├── config/              # Configuration & authentication
│   ├── cli/                 # CLI commands
│   └── mcp/                 # MCP server implementation
├── pkg/models/              # Shared types
├── test/
│   ├── fixtures/            # Test images (DO NOT MODIFY - use only in tests)
│   └── integration/         # Integration tests
└── docs/                    # Documentation
```

## Architecture Patterns

### Pure Go Philosophy
This project uses **pure Go with zero C dependencies** for maximum portability:
- Single binary distribution (no system dependencies)
- Cross-compilation to any platform (Linux, macOS, Windows, ARM)
- Uses `disintegration/imaging` library (not bimg/libvips)
- Never add C library dependencies

### Image Processing Flow
1. Load image with format auto-detection
2. Apply operation (resize/scale/crop/compress) using Lanczos resampling
3. Handle transparency (PNG→JPG uses white background)
4. Save with proper format encoding

### Configuration Hierarchy (Priority Order)
1. Command-line flags (highest priority)
2. Environment variables (`GEMINI_API_KEY`, `VERTEX_API_KEY`, `GOOGLE_APPLICATION_CREDENTIALS`)
3. Config file (`~/.gimage/config.md`)
4. Default values (lowest priority)

### API Client Pattern
Both Gemini and Vertex clients follow the same interface:
```go
type ImageGenerator interface {
    GenerateImage(prompt string, options GenerateOptions) (*GeneratedImage, error)
    ValidateCredentials() error
}
```

### Error Handling
- Return errors with context (use `fmt.Errorf` with `%w`)
- Provide actionable error messages
- Never panic in production code
- Validate inputs early

## Development Workflow

### Adding a New CLI Command
1. Create command file in `internal/cli/`
2. Implement using Cobra patterns (see existing commands)
3. Add flags with Viper binding
4. Wire up to root command in `cmd/gimage/main.go`
5. Add unit tests
6. Update documentation in `docs/API.md`

### Adding Image Processing Operations
1. Create operation file in `internal/imaging/`
2. Use `disintegration/imaging` library exclusively
3. Handle all supported formats (PNG, JPG, WebP, GIF, TIFF, BMP)
4. Add comprehensive error handling
5. Create unit tests with fixtures from `test/fixtures/`
6. Benchmark critical operations

### Testing Strategy
- Unit tests: >80% coverage required
- Integration tests: Mock external APIs (Gemini/Vertex)
- Test fixtures: Use existing images in `test/fixtures/` (DO NOT MODIFY)
- Benchmark: Profile image processing operations
- Table-driven tests for multiple scenarios

## API Integration

### Gemini API
- Use `google.golang.org/genai` SDK
- Models: `gemini-2.5-flash-image`, `gemini-2.0-flash-preview-image-generation`
- Authentication: API key via `GEMINI_API_KEY` env var or `~/.gimage/config.md`
- Setup: `gimage auth gemini` (interactive)
- Get API key from: https://aistudio.google.com/app/apikey
- Default size: 1024x1024
- Supports styles: photorealistic, artistic, anime

### Vertex AI
- Use `cloud.google.com/go/vertexai` SDK
- Models: `imagen-3.0-generate-002`, `imagen-4`
- Authentication options:
  1. **Express Mode** (recommended for dev): API key via `VERTEX_API_KEY` env var or config file
  2. **Full Mode**: Service account via `GOOGLE_APPLICATION_CREDENTIALS` env var
  3. **Full Mode**: Application Default Credentials (run `gcloud auth application-default login`)
- Setup: `gimage auth vertex` (interactive, offers all 3 modes)
- Supports 2K resolution generation
- Cost estimation required before generation

### Retry Logic
- Max 3 retry attempts with exponential backoff
- Handle rate limits gracefully
- Clear error messages for quota/permission issues

## MCP Server

The CLI can run as an MCP server for Claude integration.

### MCP Tools Exposed
1. `generate_image` - Text-to-image generation
2. `resize_image` - Resize to dimensions
3. `scale_image` - Scale by factor
4. `crop_image` - Crop region
5. `compress_image` - Compress file
6. `batch_process_images` - Batch operations
7. `get_image_info` - Image metadata

### Starting MCP Server
```bash
gimage serve
```

Configuration in `mcp-server.json` defines tool schemas.

## Authentication & Configuration

### Interactive Authentication Setup
The CLI provides interactive authentication commands:

```bash
# Setup Gemini API (simple API key)
gimage auth gemini

# Setup Vertex AI (3 modes: Express API key, Service Account, or ADC)
gimage auth vertex
```

These commands create/update `~/.gimage/config.md` with your credentials.

### Configuration File Format
Config file uses **markdown format** (not YAML/JSON) at `~/.gimage/config.md`:

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

Format: `**key**: value` on each line. Comments start with `#`.

## Security & Best Practices

### Credentials
- Never log API keys or credentials
- Config file automatically created with 0600 permissions (owner read/write only)
- Use `gimage auth` commands instead of manually editing config
- Use environment variables for CI/CD (override config file values)

### Code Quality
- Follow Go idioms and conventions
- Keep functions small and focused
- Use golangci-lint for linting
- Document all public APIs with godoc comments

### Performance
- Leverage Go's concurrency for batch operations
- Default to 4 parallel workers (configurable)
- Profile before optimizing
- Monitor memory usage for large images

## Common Development Patterns

### Loading and Saving Images
```go
// Always use imaging package
img, err := imaging.Open(inputPath)
if err != nil {
    return fmt.Errorf("failed to open image: %w", err)
}

// Process image
result := imaging.Resize(img, width, height, imaging.Lanczos)

// Save with format detection
err = imaging.Save(result, outputPath)
```

### Concurrent Processing
```go
// Use worker pool pattern for batch operations
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

## Release Process

1. Update version in code
2. Run full test suite: `make test`
3. Build all platforms: `make build-all`
4. Create git tag: `git tag v1.x.x`
5. GitHub Actions handles release automation

## Implementation Phases

Reference `IMAGE_CLI_PLAN.md` for detailed implementation prompts:
- Phase 1: Project initialization
- Phase 2: Image processing core
- Phase 3: Gemini API integration
- Phase 4: Vertex AI integration
- Phase 5: CLI commands
- Phase 6: Configuration system
- Phase 7: Testing suite
- Phase 8: Documentation
- Phase 9: MCP server
- Phase 10: Build & distribution
