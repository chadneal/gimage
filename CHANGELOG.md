# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

(empty - ready for next release)

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
