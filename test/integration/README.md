# Integration Tests

This directory contains end-to-end (E2E) integration tests for the gimage project.

## Test Categories

### 1. Production Quality Tests (`production_quality_test.go`)

Tests for the three critical production quality improvements:

#### Size Enforcement Tests
- Verifies that generated images are automatically resized to match requested dimensions
- Tests multiple aspect ratios: 512x512, 1024x1024, 512x768, 768x512
- Ensures AI models that return incorrect sizes are automatically corrected
- **Costs**: Uses Gemini API (free tier available)

#### Format Conversion Tests
- Verifies automatic format conversion to requested output format
- Tests all supported formats: PNG, JPEG, WebP, GIF, BMP, TIFF
- Ensures generated images can be read by standard image libraries
- **Costs**: Uses Gemini API (free tier available)

#### Output Path Handling Tests
- Verifies all CLI commands correctly handle various output path formats
- Tests absolute paths, relative paths, paths with spaces, nested directories
- Tests all commands: resize, scale, crop, convert
- **Costs**: FREE (uses local test fixtures only)

### 2. API E2E Tests (`generate_e2e_test.go`)

Tests for real API integration:
- Gemini API tests
- Vertex AI tests
- AWS Bedrock tests
- **Costs**: Variable based on API pricing

### 3. CLI E2E Tests (`cli_e2e_test.go`)

Tests for CLI commands using local fixtures:
- Resize, scale, crop commands
- **Costs**: FREE

### 4. WebP Tests (`webp_test.go`)

Tests for WebP format support:
- **Costs**: FREE

## Running Tests

### Run All Tests (including E2E)
```bash
make test
```

### Run Only Unit Tests (no API calls)
```bash
go test ./... -short
```

### Run Only Integration Tests
```bash
go test -tags=e2e ./test/integration/...
```

### Run Specific Test Suite

#### Size Enforcement Tests
```bash
go test -tags=e2e ./test/integration/ -run TestSizeEnforcement -v
```

#### Format Conversion Tests
```bash
go test -tags=e2e ./test/integration/ -run TestFormatConversion -v
```

#### Output Path Handling Tests
```bash
go test -tags=e2e ./test/integration/ -run TestOutputPathHandling -v
```

## Test Fixtures

Test fixtures are located in `test/fixtures/`:
- `test_image.png` (800x600) - General testing
- `test_image_512x512.png` (512x512) - Size enforcement testing
- `small_test.png` (200x150) - Small image testing

To regenerate fixtures:
```bash
cd test/fixtures
go run generate_fixtures.go
```

## Cost Estimates

### Free Tests
- Output path handling: FREE (local only)
- CLI E2E tests: FREE (local only)
- WebP tests: FREE (local only)

### Paid Tests
- Size enforcement (4 test cases): ~$0.00 (Gemini free tier)
- Format conversion (6 test cases): ~$0.00 (Gemini free tier)
- Vertex AI tests: ~$0.02 per test
- AWS Bedrock tests: ~$0.04 per test

**Note**: Gemini has a free tier of 1500 requests/day, so most tests are free.

## Authentication

Tests require API credentials to be configured:

### Gemini API (Free Tier)
```bash
export GEMINI_API_KEY="your-key"
# or
gimage auth gemini
```

### Vertex AI (Paid)
```bash
export VERTEX_API_KEY="your-key"
export VERTEX_PROJECT="your-project-id"
# or
gimage auth vertex
```

### AWS Bedrock (Paid)
```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"
# or
gimage auth bedrock
```

## Skipping Tests

Tests will automatically skip if credentials are not available:
- Size enforcement: Skips if no Gemini credentials
- Format conversion: Skips if no Gemini credentials
- Output path handling: Never skips (uses local fixtures)
- API E2E tests: Skip if respective API credentials not available

## Test Output

Example successful output:
```
=== RUN   TestSizeEnforcement
=== RUN   TestSizeEnforcement/Square_512x512
    production_quality_test.go:45: Generating image with requested size: 512x512
    production_quality_test.go:68: Requested: 512x512, Got: 512x512
=== RUN   TestSizeEnforcement/Square_1024x1024
    production_quality_test.go:45: Generating image with requested size: 1024x1024
    production_quality_test.go:68: Requested: 1024x1024, Got: 1024x1024
--- PASS: TestSizeEnforcement (5.23s)
```

## Troubleshooting

### "Test skipped" messages
- Ensure credentials are configured
- Check environment variables or config file

### "Rate limit exceeded" errors
- Gemini: Wait for quota reset (1500/day limit)
- Vertex/Bedrock: Check your account limits

### "Output file not created" errors
- Check file permissions
- Verify parent directory exists
- Check disk space

## Contributing

When adding new tests:
1. Use build tag `// +build e2e` for tests that cost money
2. Skip tests if credentials not available
3. Document cost estimates
4. Use Gemini API for cheaper/free tests when possible
5. Clean up temp files using `t.TempDir()`
