# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

(empty - ready for next release)

## [1.1.46] - 2025-11-05

```markdown
## [Unreleased]

### Added
- Interactive TUI (Terminal User Interface) with main menu, batch processing, generation flow, and settings management
- Batch operation history tracking with persistent storage
- Progress reporter for real-time operation feedback
- Production quality test suite with comprehensive integration tests
- Image compression operation with quality control
- TUI documentation and feature tour
- Test fixtures (small_test.png, test_image.png, test_image_512x512.png)

### Changed
- Simplified CLI command outputs for better TUI integration
- Improved image processing operations (resize, scale, crop, convert) with enhanced error handling
- Streamlined documentation: consolidated guides into concise references
- Reduced project documentation by 56% (removed planning and implementation tracking docs)
- Updated lambda.md from 1,385 to 272 lines (removed CDK code, kept deployment guide)
- Updated INTEGRATION_GUIDE.md to focus on crisp examples only

### Removed
- Project planning documents (RELEASING.md, roadmap.md, HOMEBREW.md)
- Implementation tracking docs (DEPLOYMENT_CHECKLIST.md, LAMBDA_STATUS.md)
- Research/analysis docs (MCP_LLM_LEARNING_ANALYSIS.md, AI_TOOL_CALLING_IMPROVEMENTS.md, AWS_BEDROCK_SDK_GUIDE.md, etc.)
- Redundant documentation (API_REFERENCE.md, SWAGGER_SETUP.md, RELEASE_NOTES.md, etc.)
```


## [1.1.43] - 2025-11-02

### Added

- MCP Prompts feature: New prompt templates for image generation, batch processing, and common workflows
- MCP server now exposes 13 prompt templates via the prompts/list capability
- Comprehensive documentation for MCP Prompts design and implementation (MCP_PROMPTS_DESIGN.md, MCP_PROMPTS_IMPLEMENTATION.md)
- Analysis documentation for LLM learning patterns with MCP (MCP_LLM_LEARNING_ANALYSIS.md)

### Changed

- Enhanced MCP tool descriptions with more actionable guidance for LLM clients
- Improved MCP handler with prompt list and get capabilities
- Updated MCP server to register prompt templates on initialization


## [1.1.41] - 2025-11-02

### Changed
- 


## [1.1.40] - 2025-11-02

### Changed
- Build number incremented to 1.1.40 (automatic versioning from git commit count)


## [1.1.38] - 2025-11-02

### Changed
- Upgraded GoReleaser configuration to v2 format for improved build and release automation


## [1.1.36] - 2025-11-02

### Changed
- Build number incremented to 1.1.36 (automatic versioning from git commit count)


## [1.1.34] - 2025-11-02

### Changed
- 


## [1.1.33] - 2025-11-02

### Changed
- Updated .gitignore patterns for improved exclusion rules


## [1.1.32] - 2025-11-02

### Added

- AWS Bedrock Nova Canvas integration with dual implementation modes (REST and SDK)
- AWS Bedrock authentication setup via `gimage auth bedrock` command
- Nova Canvas model support (`amazon.nova-canvas-v1:0`) with quality presets (standard/premium)
- Advanced generation controls: negative prompts, CFG scale, seed, and quality settings
- Comprehensive AWS Bedrock documentation (SDK guide, quickstart, onboarding guide)
- Testing infrastructure with coverage reporting tools (`cmd/coverage-report`, `cmd/test-report`, `cmd/test-summary`)
- Extensive test suites for Bedrock REST and SDK clients (382+ and 305+ test cases respectively)
- MCP tools test coverage (batch, convert, generate operations)
- End-to-end integration tests for CLI and generation workflows
- Testing best practices documentation (TESTING.md)
- Model onboarding guide (MODEL_ONBOARDING.md) for adding new backends
- Documentation index (DOCUMENTATION_INDEX.md) for centralized reference
- Coverage report scripts with detailed HTML output

### Changed

- Updated CLAUDE.md with multi-backend architecture guidance and AWS Bedrock sections
- Enhanced `gimage generate` command with AWS-specific flags (quality, seed, CFG scale, negative prompts)
- Expanded configuration system to support AWS credentials and region settings
- Updated README.md with AWS Bedrock usage examples
- Improved MCP_TOOLS.md with Bedrock integration examples
- Enhanced Makefile with test coverage and reporting targets
- Refactored generate models to support backend-specific options
- Updated auth.go with Bedrock credential management (REST and SDK modes)

### Fixed

- Image scaling operations with improved precision
- Crop and scale CLI commands with better error handling


## [1.1.29] - 2025-11-02

### Changed
- 


## [1.1.28] - 2025-11-02

### Changed
- 


## [1.1.27] - 2025-11-02

### Changed
- 


## [1.1.26] - 2025-11-02

### Removed
- Removed MCP tool tests (convert_test.go, generate_test.go, models_test.go) that were incompatible with current implementation


## [1.1.23] - 2025-11-02

### Added
- Comprehensive model pricing and announcement system with cost tracking and latest model information
- Unit tests for generate command with coverage for both Gemini and Vertex AI backends
- Unit tests for convert operation with format conversion validation
- Unit tests for resize operation with comprehensive dimension and format testing
- Unit tests for crop operation with boundary and validation testing
- Automated changelog update script for release process

### Changed
- Enhanced generate command with model pricing display and cost estimation
- Improved MCP server with model information and pricing details
- Updated RELEASING.md with streamlined release workflow and automation improvements
- Refactored Makefile with improved test coverage reporting and build targets


## [1.1.19] - 2025-11-02

### Changed
- Build number incremented to 1.1.19 (automatic versioning from git commit count)


## [1.1.18] - 2025-11-01

### Changed
- Build number incremented to 1.1.18 (automatic versioning from git commit count)


## [1.1.9] - 2025-11-01

### Added
- **Automated version synchronization** between CLI and npm package
- **Build number versioning** using git commit count (format: 1.1.[build])
- WebP support via nativewebp library (pure Go, zero C dependencies)
- CLI `convert` command for format conversion
- Comprehensive integration tests for WebP
- End-to-end tests for all 10 MCP tools
- Help text displayed when running `gimage` with no arguments
- Complete release automation with GoReleaser
- GitHub Actions workflows for CI and releases
- npm package for MCP server distribution
- Homebrew tap for macOS/Linux distribution
- Comprehensive RELEASING.md guide
- `make version` and `make sync-version` commands

### Changed
- **Version numbering scheme** to 1.1.[commit_count] for automatic sync
- Root command now shows help instead of crashing when run without arguments
- All MCP tools now support WebP output format
- Homebrew tap repository renamed to `homebrew-tap` (conventional naming)
- Documentation updated for new distribution methods

### Fixed
- Root command exit behavior
- WebP encoding in all contexts (CLI, MCP server, programmatic usage)
- Version synchronization between CLI binary and npm package

## [0.1.1] - 2025-11-01

### Added
- Automatic format conversion based on output file extension
  - Specify `-o output.jpg` to save as JPEG
  - Specify `-o output.png` to save as PNG
  - Specify `-o output.gif` to save as GIF
  - Specify `-o output.bmp` to save as BMP
  - Specify `-o output.tiff` to save as TIFF
- Intelligent transparency handling (converts transparent areas to white when saving to formats that don't support transparency)
- Format normalization (automatically handles .jpg vs .jpeg, .tif vs .tiff)

### Changed
- SaveImage now automatically converts image format based on file extension

## [0.1.0] - 2025-11-01

### Added
- Initial release of gimage CLI tool
- AI-powered image generation using Google Gemini 2.5 Flash Image
- AI-powered image generation using Vertex AI Imagen 4
- Image processing operations:
  - Resize: Change image dimensions
  - Scale: Scale by percentage
  - Crop: Extract specific regions
  - Compress: Reduce file size
- Batch processing with concurrent operations
- MCP server for Claude integration
- Support for multiple image formats: PNG, JPG, WebP, GIF, TIFF, BMP
- Pure Go implementation with zero C dependencies
- Cross-platform support (Linux, macOS, Windows, ARM)
- Interactive authentication setup:
  - `gimage auth gemini` - Gemini API key setup
  - `gimage auth vertex` - Vertex AI setup (Express Mode or Full Mode)
- Smart credential detection - auto-selects API based on available credentials
- Configuration system with markdown-based config file (~/.gimage/config.md)
- Comprehensive CLI with Cobra framework

### Features
- Text-to-image generation with customizable prompts
- Multiple generation styles: photorealistic, artistic, anime
- Configurable image sizes and aspect ratios
- Negative prompts for image generation
- Seed support for reproducible results
- Verbose mode for debugging
- Model listing and auto-detection
- Express Mode for Vertex AI (API key authentication)
- Full Mode for Vertex AI (service account authentication)

### Technical
- Built with Go 1.22+
- Uses disintegration/imaging library for image processing
- Gemini API integration via REST
- Vertex AI integration via REST (Express Mode) and SDK (Full Mode)
- Concurrent batch processing with worker pools
- Comprehensive error handling and validation
