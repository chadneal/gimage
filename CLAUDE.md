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

## API Integration - Multi-Backend Architecture

**IMPORTANT**: Gimage supports multiple AI generation backends with both SDK and REST API implementations. Each backend has its own client type but shares a common interface pattern.

### Architecture Overview

```
Image Generation Backends:
├── Gemini API (REST)        -> generate.NewGeminiRESTClient(apiKey)
├── Vertex AI (Express Mode) -> generate.NewVertexRESTClient(apiKey, project, location)  [REST]
├── Vertex AI (Full Mode)    -> generate.NewVertexSDKClient(ctx, project, location)      [SDK]
└── Future: Bedrock, Nova, etc.
```

**Common Client Interface Pattern:**
All clients implement these methods:
```go
GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error)
Close() error  // Cleanup resources
```

### Backend Selection Logic (Priority Order)
1. **Explicit flag**: `--api gemini` or `--api vertex`
2. **Auto-detect from model**: Model name implies backend (e.g., "imagen-4" → vertex)
3. **Auto-detect from credentials**: Check which API keys are configured
4. **Config default**: Use `default_api` from `~/.gimage/config.md`
5. **Fallback**: Default to Gemini if both are available

### Gemini API Backend

**Implementation**: REST API client (`gemini_rest.go`)

**Setup**:
```bash
gimage auth gemini
```

**Models**:
- `gemini-2.5-flash-image` (default, recommended)
- `gemini-2.0-flash-preview-image-generation`

**Authentication**:
- API key via `GEMINI_API_KEY` env var or `~/.gimage/config.md`
- Get free API key: https://aistudio.google.com/app/apikey
- Free tier: 1500 requests/day

**Usage in Code**:
```go
import "github.com/apresai/gimage/internal/generate"
import "github.com/apresai/gimage/internal/config"

// Get API key from config or env
key, err := config.GetGeminiAPIKey("")
client, err := generate.NewGeminiRESTClient(key)
defer client.Close()

ctx := context.Background()
options := models.GenerateOptions{
    Model: "gemini-2.5-flash-image",
    Size: "1024x1024",
    Style: "photorealistic",
}
img, err := client.GenerateImage(ctx, prompt, options)
```

### Vertex AI Backend

**Two Implementation Modes**:
1. **Express Mode** - REST API with API key (simpler, recommended for dev)
2. **Full Mode** - SDK with service account or ADC (production-grade)

**Setup**:
```bash
gimage auth vertex  # Interactive wizard offers both modes
```

**Models**:
- `imagen-3.0-generate-002` (Imagen 3)
- `imagen-4` (latest, highest quality, up to 2048x2048)

#### Express Mode (REST API)

**Implementation**: `vertex_rest.go`

**Authentication**:
- API key via `VERTEX_API_KEY` env var or config
- Requires `VERTEX_PROJECT` and `VERTEX_LOCATION`

**Usage**:
```go
apiKey, err := config.GetVertexAPIKey("")
project := os.Getenv("VERTEX_PROJECT") // or from config
location := "us-central1"

client, err := generate.NewVertexRESTClient(apiKey, project, location)
defer client.Close()

img, err := client.GenerateImage(ctx, prompt, options)
```

#### Full Mode (SDK)

**Implementation**: `vertex_sdk.go`

**Authentication Options**:
1. Service account JSON via `GOOGLE_APPLICATION_CREDENTIALS`
2. Application Default Credentials (`gcloud auth application-default login`)

**Usage**:
```go
project := os.Getenv("VERTEX_PROJECT")
location := "us-central1"

client, err := generate.NewVertexSDKClient(ctx, project, location)
defer client.Close()

img, err := client.GenerateImage(ctx, prompt, options)
```

### Future Backends

When adding new backends (Bedrock, Nova, etc.):

1. **Create new client file**: `internal/generate/bedrock_sdk.go` or `bedrock_rest.go`
2. **Implement common interface**:
   ```go
   func NewBedrockClient(ctx, cfg) (*BedrockClient, error)
   func (c *BedrockClient) GenerateImage(ctx, prompt, options) (*GeneratedImage, error)
   func (c *BedrockClient) Close() error
   ```
3. **Add to model detection**: Update `generate.DetectAPIFromModel()`
4. **Add auth setup**: Create `gimage auth bedrock` command
5. **Update docs**: Add to this section and MCP_TOOLS.md
6. **Add tests**: Create `bedrock_test.go` for unit tests (request building, response parsing, validation)
7. **Integration tests**: Add `bedrock_integration_test.go` with real API calls (manual only)

### Testing Strategy for Multi-Backend

**CRITICAL: DO NOT MOCK CLOUD PROVIDER APIs**

Mocking cloud provider APIs (AWS, Google, etc.) provides **zero value** and creates a false sense of security. Cloud providers change their APIs, error formats, rate limits, and behaviors regularly. Mocks become stale immediately and don't catch real-world issues.

**What TO Test**:
- ✅ Request payload building (validate JSON structure before sending)
- ✅ Response parsing logic (with real example responses from docs)
- ✅ Error message formatting and actionability
- ✅ Input validation (dimensions, prompts, parameters)
- ✅ Configuration loading and credential detection
- ✅ CLI flag parsing and option handling

**What NOT TO Test**:
- ❌ Mocked API calls (worthless, creates false confidence)
- ❌ Fake HTTP responses (providers change formats)
- ❌ Simulated errors (real errors differ from assumptions)

**Testing Approach**:

1. **Unit Tests** - Test everything EXCEPT the actual API call:
```go
// ✅ GOOD: Test request building
func TestBuildNovaCanvasRequest(t *testing.T) {
    request := buildNovaCanvasRequest("test prompt", GenerateOptions{
        Size: "1024x1024",
        Seed: 42,
    })

    // Validate the JSON structure matches AWS docs
    assert.Equal(t, "TEXT_IMAGE", request.TaskType)
    assert.Equal(t, 1024, request.ImageGenerationConfig.Width)
    assert.Equal(t, 42, request.ImageGenerationConfig.Seed)
}

// ✅ GOOD: Test response parsing with real example
func TestParseNovaCanvasResponse(t *testing.T) {
    // Use actual response from AWS documentation
    realResponse := `{"images": ["iVBORw0KGgo..."], "error": null}`

    result, err := parseNovaCanvasResponse([]byte(realResponse))
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Images)
}

// ❌ BAD: Mocking the AWS SDK
// type mockBedrockClient struct { ... }  // DON'T DO THIS
```

2. **Integration Tests** - Test against REAL APIs (manual only):
```go
// +build integration

// These tests cost money and require real credentials
// Run manually: go test -tags=integration ./internal/generate/...
func TestBedrockRealAPI(t *testing.T) {
    if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
        t.Skip("Real AWS credentials not configured")
    }

    client, _ := generate.NewBedrockSDKClient(ctx, "us-east-1")

    // Real API call (costs $0.04)
    img, err := client.GenerateImage(ctx, "simple test", GenerateOptions{})
    assert.NoError(t, err)
    assert.NotEmpty(t, img.Data)
}
```

3. **Manual Testing** - Primary validation method:
```bash
# Test with real credentials (costs ~$0.04-0.20)
gimage generate --api bedrock "test image" --verbose
gimage generate --api vertex "test image" --verbose
gimage generate --api gemini "test image" --verbose
```

**Why This Approach Works**:
- Unit tests catch logic bugs without API calls
- Integration tests validate real behavior (run manually before releases)
- Manual testing is quick and catches UX issues
- No maintenance burden of brittle mocks
- Tests accurately reflect production behavior

**Test File Structure**:
```
internal/generate/
├── bedrock_sdk.go          # Implementation
├── bedrock_request.go      # Request building (unit testable)
├── bedrock_response.go     # Response parsing (unit testable)
├── bedrock_test.go         # Unit tests (no API calls)
└── bedrock_integration_test.go  # Real API tests (manual only)
```

### Configuration File Support

All backends read from `~/.gimage/config.md`:
```markdown
# Gemini Configuration
**gemini_api_key**: AIzaSy...

# Vertex AI Configuration
**vertex_api_key**: AIzaSy...          # For Express Mode
**vertex_project**: my-gcp-project
**vertex_location**: us-central1
**vertex_credentials_path**: ~/.gimage/credentials/sa.json  # For Full Mode

# Default Backend
**default_api**: gemini  # or "vertex"
```

### Retry Logic (All Backends)
- Max 3 retry attempts with exponential backoff
- Initial backoff: 1 second, max: 10 seconds
- Retryable errors: rate limits, timeouts, 503 errors
- Non-retryable: invalid key, bad params, permission denied

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

### Documentation and Dates
- **ALWAYS use the system `date` command to get the current date** when creating or updating documentation
- Never hardcode dates in documentation - they become outdated immediately
- Use `date +%Y-%m-%d` for YYYY-MM-DD format (ISO 8601 standard)
- When updating CHANGELOG.md, RELEASING.md, or any documentation with dates, run the date command first

Example workflow:
```bash
# Get current date for documentation
date +%Y-%m-%d
# Output: 2025-11-01

# Use this date in CHANGELOG.md entries
## [0.2.0] - 2025-11-01
```

**Why this matters**: Hardcoded dates quickly become incorrect and make documentation confusing. Always fetch the current system date dynamically.

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

## Git Usage Policy

**IMPORTANT**: Do NOT use git commands (commit, push, tag, etc.) unless the user explicitly asks for it.

**Examples of when NOT to use git**:
- After creating or modifying files
- After completing a feature implementation
- After fixing bugs or making improvements
- When you think "this should be committed"

**Examples of when TO use git**:
- User says "commit this"
- User says "push to GitHub"
- User says "create a git tag"
- User explicitly requests git operations in their message

**Why**: The user controls when and how code is committed. Automatic commits can:
- Interrupt their workflow
- Create unwanted commit history
- Commit incomplete or experimental changes
- Bypass their review process

If you complete work and think it should be committed, simply inform the user what was done and let them decide whether to commit.

## Release Process

1. Update version in code
2. Run full test suite: `make test`
3. Build all platforms: `make build-all`
4. Create git tag: `git tag v1.x.x` (only when user requests)
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
