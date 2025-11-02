# AWS Bedrock SDK Implementation Guide for Nova Canvas

**Date**: 2025-11-02
**Based on**: AWS Bedrock official documentation and AWS SDK for Go v2

This guide provides detailed technical implementation details for integrating AWS Bedrock Nova Canvas using the Go SDK, based on official AWS documentation.

---

## Table of Contents

1. [AWS Bedrock Nova Canvas Overview](#aws-bedrock-nova-canvas-overview)
2. [Request/Response Format](#requestresponse-format)
3. [Go SDK Implementation](#go-sdk-implementation)
4. [Authentication Methods](#authentication-methods)
5. [Error Handling](#error-handling)
6. [Code Examples](#code-examples)
7. [Best Practices](#best-practices)

---

## AWS Bedrock Nova Canvas Overview

### Model Information

**Model ID**: `amazon.nova-canvas-v1:0`

**Capabilities**:
- Text-to-image generation
- Supports styles and quality settings
- Configurable seeds for reproducibility
- Maximum resolution: 2048x2048 pixels
- Supported aspect ratios: 1:1, 16:9, 9:16, 4:3, 3:4

**Pricing**:
- Standard quality: $0.04 per image
- Premium quality: $0.08 per image

**Rate Limits**:
- 10 requests per second per account
- Region-specific (check AWS documentation for updates)

**Available Regions**:
- us-east-1 (US East - N. Virginia)
- us-west-2 (US West - Oregon)
- Check AWS docs for latest regional availability

---

## Request/Response Format

### Request Structure

The Nova Canvas model expects a JSON payload with this structure:

```json
{
  "taskType": "TEXT_IMAGE",
  "textToImageParams": {
    "text": "A stylized picture of a cute old steampunk robot",
    "negativeText": "low quality, blurry, distorted"  // Optional
  },
  "imageGenerationConfig": {
    "numberOfImages": 1,
    "quality": "standard",  // or "premium"
    "height": 1024,
    "width": 1024,
    "cfgScale": 7.0,  // Guidance scale (1.0-20.0)
    "seed": 42  // Optional, 0-858,993,459
  }
}
```

### Request Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `taskType` | string | Yes | - | Must be "TEXT_IMAGE" |
| `textToImageParams.text` | string | Yes | - | Prompt describing the image |
| `textToImageParams.negativeText` | string | No | - | What to avoid in the image |
| `imageGenerationConfig.numberOfImages` | integer | No | 1 | Number of images (1-5) |
| `imageGenerationConfig.quality` | string | No | "standard" | "standard" or "premium" |
| `imageGenerationConfig.height` | integer | No | 1024 | Height in pixels (512-2048) |
| `imageGenerationConfig.width` | integer | No | 1024 | Width in pixels (512-2048) |
| `imageGenerationConfig.cfgScale` | float | No | 7.0 | Guidance scale (1.0-20.0) |
| `imageGenerationConfig.seed` | integer | No | random | Seed for reproducibility (0-858,993,459) |

**Note**: Width and height must result in a total pixel count between 262,144 (512x512) and 4,194,304 (2048x2048).

### Response Structure

```json
{
  "images": [
    "iVBORw0KGgoAAAANSUhEUgAA..."  // Base64-encoded PNG image data
  ],
  "error": null  // or error message if generation failed
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `images` | array[string] | Array of base64-encoded PNG images |
| `error` | string or null | Error message if generation failed |

---

## Go SDK Implementation

### Required Dependencies

```bash
# Install AWS SDK v2 for Go
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime
go get github.com/aws/aws-sdk-go-v2/aws
```

### Package Structure

```go
package generate

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
    "github.com/apresai/gimage/pkg/models"
    "github.com/sony/gobreaker"
    "github.com/spf13/viper"
)
```

### Client Structure

```go
// BedrockSDKClient uses AWS Bedrock Runtime for image generation
type BedrockSDKClient struct {
    client         *bedrockruntime.Client
    region         string
    verbose        bool
    circuitBreaker *gobreaker.CircuitBreaker
}
```

### Constructor Implementation

```go
// NewBedrockSDKClient creates a new AWS Bedrock SDK client
func NewBedrockSDKClient(ctx context.Context, region string) (*BedrockSDKClient, error) {
    // Default region if not provided
    if region == "" {
        region = os.Getenv("AWS_REGION")
        if region == "" {
            region = "us-east-1"
        }
    }

    // Load AWS SDK configuration
    // This automatically handles:
    // - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    // - Shared credentials file (~/.aws/credentials)
    // - IAM role credentials (EC2, ECS, Lambda)
    // - AWS SSO profiles
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion(region),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    // Create Bedrock Runtime client
    client := bedrockruntime.NewFromConfig(cfg)

    // Check verbose mode
    verbose := viper.GetBool("verbose") ||
               os.Getenv("GIMAGE_VERBOSE") == "true" ||
               os.Getenv("VERBOSE") == "true"

    return &BedrockSDKClient{
        client:         client,
        region:         region,
        verbose:        verbose,
        circuitBreaker: newCircuitBreaker("BedrockAPI"),
    }, nil
}
```

### Request/Response Types

```go
// NovaCanvasRequest represents the Nova Canvas API request format
type NovaCanvasRequest struct {
    TaskType          string                     `json:"taskType"`
    TextToImageParams NovaCanvasTextToImageParams `json:"textToImageParams"`
    ImageGenerationConfig NovaCanvasImageConfig `json:"imageGenerationConfig"`
}

type NovaCanvasTextToImageParams struct {
    Text         string `json:"text"`
    NegativeText string `json:"negativeText,omitempty"`
}

type NovaCanvasImageConfig struct {
    NumberOfImages int     `json:"numberOfImages"`
    Quality        string  `json:"quality"`
    Height         int     `json:"height"`
    Width          int     `json:"width"`
    CfgScale       float64 `json:"cfgScale"`
    Seed           int     `json:"seed,omitempty"`
}

// NovaCanvasResponse represents the Nova Canvas API response format
type NovaCanvasResponse struct {
    Images []string `json:"images"` // Base64-encoded images
    Error  string   `json:"error,omitempty"`
}
```

### GenerateImage Implementation

```go
func (c *BedrockSDKClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    // Validate prompt
    if err := ValidatePrompt(prompt); err != nil {
        return nil, err
    }

    // Enhance prompt for better results
    enhancedPrompt := EnhancePrompt(prompt)

    // Use custom model if provided, otherwise default
    modelID := "amazon.nova-canvas-v1:0"
    if options.Model != "" {
        modelID = options.Model
    }

    // Generate image with circuit breaker and retry logic
    var lastErr error
    backoff := retryBackoffInitial

    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Execute through circuit breaker
        result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
            return c.generateWithRetry(ctx, modelID, enhancedPrompt, options)
        })

        if err == nil {
            return result.(*models.GeneratedImage), nil
        }

        lastErr = err

        // Check if circuit breaker is open
        if isCircuitBreakerError(err) {
            c.logVerbose("Circuit breaker is open, failing fast")
            return nil, fmt.Errorf("API circuit breaker is open (too many failures): %w", err)
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            return nil, err
        }

        // Exponential backoff
        if attempt < maxRetries {
            select {
            case <-ctx.Done():
                return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
            case <-time.After(backoff):
                backoff *= 2
                if backoff > retryBackoffMax {
                    backoff = retryBackoffMax
                }
            }
        }
    }

    return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
```

### Core Generation Logic

```go
func (c *BedrockSDKClient) generateWithRetry(ctx context.Context, modelID, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    // Parse dimensions from size string (e.g., "1024x1024")
    width, height := parseDimensions(options.Size)

    c.logVerbose("Building request for model: %s", modelID)
    c.logVerbose("Requested dimensions: %dx%d", width, height)
    c.logVerbose("Prompt: %s", prompt)

    // Determine quality level
    quality := "standard"
    if options.Style == "premium" || options.Style == "high-quality" {
        quality = "premium"
    }

    // Build request payload
    request := NovaCanvasRequest{
        TaskType: "TEXT_IMAGE",
        TextToImageParams: NovaCanvasTextToImageParams{
            Text:         prompt,
            NegativeText: options.NegativePrompt,
        },
        ImageGenerationConfig: NovaCanvasImageConfig{
            NumberOfImages: 1,
            Quality:        quality,
            Height:         height,
            Width:          width,
            CfgScale:       7.0, // Default guidance scale
            Seed:           int(options.Seed),
        },
    }

    // Marshal to JSON
    requestJSON, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    c.logVerbose("Request body: %s", string(requestJSON))

    // Invoke model using AWS SDK
    c.logVerbose("Invoking model: %s in region %s", modelID, c.region)

    input := &bedrockruntime.InvokeModelInput{
        ModelId:     aws.String(modelID),
        ContentType: aws.String("application/json"),
        Accept:      aws.String("application/json"),
        Body:        requestJSON,
    }

    response, err := c.client.InvokeModel(ctx, input)
    if err != nil {
        c.logVerbose("InvokeModel failed: %v", err)
        return nil, c.handleSDKError(err)
    }

    c.logVerbose("Response received, status code: %d", response.ResultMetadata)

    // Parse response
    var novaResponse NovaCanvasResponse
    if err := json.Unmarshal(response.Body, &novaResponse); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    // Check for API errors in response
    if novaResponse.Error != "" {
        return nil, fmt.Errorf("API error: %s", novaResponse.Error)
    }

    if len(novaResponse.Images) == 0 {
        return nil, fmt.Errorf("no image generated")
    }

    // Decode base64 image
    imageData, err := base64.StdEncoding.DecodeString(novaResponse.Images[0])
    if err != nil {
        return nil, fmt.Errorf("failed to decode base64 image: %w", err)
    }

    c.logVerbose("Successfully generated image: %d bytes", len(imageData))

    return &models.GeneratedImage{
        Data:   imageData,
        Format: "png", // Nova Canvas returns PNG
        Width:  width,
        Height: height,
        Metadata: map[string]string{
            "model":   modelID,
            "prompt":  prompt,
            "quality": quality,
            "api":     "bedrock",
            "region":  c.region,
        },
    }, nil
}
```

### Verbose Logging

```go
func (c *BedrockSDKClient) logVerbose(format string, args ...interface{}) {
    if c.verbose {
        fmt.Fprintf(os.Stderr, "[BEDROCK-SDK] "+format+"\n", args...)
    }
}
```

### Close Method

```go
func (c *BedrockSDKClient) Close() error {
    // AWS SDK clients don't require explicit closing
    return nil
}
```

---

## Authentication Methods

### Method 1: Environment Variables (Recommended for Development)

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"
```

**Pros**: Simple, works everywhere
**Cons**: Credentials in environment, not ideal for production

### Method 2: Shared Credentials File (Recommended for Local Development)

**File**: `~/.aws/credentials`

```ini
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[production]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

**File**: `~/.aws/config`

```ini
[default]
region = us-east-1

[profile production]
region = us-west-2
```

**Usage**:
```bash
# Use default profile
gimage generate "test"

# Use specific profile
AWS_PROFILE=production gimage generate "test"
```

**Pros**: Supports multiple profiles, secure file permissions
**Cons**: Requires AWS CLI setup

### Method 3: IAM Role (Recommended for Production - EC2/ECS/Lambda)

For applications running on AWS infrastructure, use IAM roles:

```go
// No credentials needed - SDK auto-detects IAM role
cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRegion("us-east-1"),
)
```

**Pros**: No credentials to manage, auto-rotated, most secure
**Cons**: Only works on AWS infrastructure

### Method 4: AWS SSO (Recommended for Enterprise)

```bash
# Configure SSO
aws configure sso

# Login
aws sso login --profile my-sso-profile

# Use with gimage
AWS_PROFILE=my-sso-profile gimage generate "test"
```

**Pros**: Enterprise-grade, MFA support, temporary credentials
**Cons**: Requires AWS CLI v2, setup complexity

### Authentication Priority (AWS SDK Default Behavior)

The SDK checks credentials in this order:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. Shared config file (`~/.aws/config`)
4. IAM role for EC2/ECS/Lambda
5. AWS SSO

---

## Error Handling

### AWS SDK Error Types

```go
func (c *BedrockSDKClient) handleSDKError(err error) error {
    if err == nil {
        return nil
    }

    errStr := err.Error()

    // AccessDeniedException - IAM permissions issue
    if contains(errStr, "AccessDeniedException") {
        return fmt.Errorf("permission denied: check IAM permissions for bedrock:InvokeModel. "+
            "Your IAM user/role needs the AmazonBedrockFullAccess policy or "+
            "a custom policy with bedrock:InvokeModel action: %w", err)
    }

    // ResourceNotFoundException - Model not found or region issue
    if contains(errStr, "ResourceNotFoundException") {
        return fmt.Errorf("model not found: verify model ID '%s' is correct and "+
            "available in region '%s'. Check model access in Bedrock console: %w",
            "amazon.nova-canvas-v1:0", c.region, err)
    }

    // ThrottlingException - Rate limit exceeded
    if contains(errStr, "ThrottlingException") {
        return fmt.Errorf("rate limit exceeded: too many requests to Bedrock. "+
            "Current limit: 10 req/sec. Wait before retrying: %w", err)
    }

    // ServiceQuotaExceededException - Account quota exceeded
    if contains(errStr, "ServiceQuotaExceededException") {
        return fmt.Errorf("service quota exceeded: check your AWS account limits "+
            "for Bedrock in the Service Quotas console: %w", err)
    }

    // ServiceUnavailableException - AWS service issue
    if contains(errStr, "ServiceUnavailableException") {
        return fmt.Errorf("service unavailable: AWS Bedrock is experiencing issues. "+
            "Check AWS status page and try again later: %w", err)
    }

    // ValidationException - Invalid request parameters
    if contains(errStr, "ValidationException") {
        return fmt.Errorf("validation error: request parameters are invalid. "+
            "Check image dimensions (512-2048), quality setting, etc.: %w", err)
    }

    // ModelNotReadyException - Model is loading
    if contains(errStr, "ModelNotReadyException") {
        return fmt.Errorf("model not ready: the model is still loading. "+
            "This is rare - try again in a few seconds: %w", err)
    }

    // ModelStreamErrorException - Streaming error (shouldn't happen for invoke_model)
    if contains(errStr, "ModelStreamErrorException") {
        return fmt.Errorf("model streaming error: %w", err)
    }

    // Generic AWS error
    return fmt.Errorf("AWS Bedrock error: %w", err)
}

// Helper function
func contains(s, substr string) bool {
    return len(s) >= len(substr) &&
           (s == substr || len(s) > len(substr) &&
            (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
             len(s) > len(substr)*2))
}
```

### Retryable Errors

```go
func isRetryableError(err error) bool {
    if err == nil {
        return false
    }

    errStr := err.Error()

    // Retry on these errors
    retryable := []string{
        "ThrottlingException",
        "ServiceUnavailableException",
        "TooManyRequestsException",
        "RequestTimeout",
        "RequestTimeoutException",
        "500", "502", "503", "504", // HTTP status codes
    }

    for _, pattern := range retryable {
        if contains(errStr, pattern) {
            return true
        }
    }

    return false
}
```

---

## Code Examples

### Example 1: Basic Image Generation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/apresai/gimage/internal/generate"
    "github.com/apresai/gimage/pkg/models"
)

func main() {
    ctx := context.Background()

    // Create client (auto-detects credentials)
    client, err := generate.NewBedrockSDKClient(ctx, "us-east-1")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // Generate image
    options := models.GenerateOptions{
        Model: "amazon.nova-canvas-v1:0",
        Size:  "1024x1024",
    }

    img, err := client.GenerateImage(ctx, "a sunset over mountains", options)
    if err != nil {
        log.Fatalf("Failed to generate image: %v", err)
    }

    // Save image
    err = os.WriteFile("output.png", img.Data, 0644)
    if err != nil {
        log.Fatalf("Failed to save image: %v", err)
    }

    fmt.Printf("Image saved: %d bytes\n", len(img.Data))
}
```

### Example 2: Premium Quality with Negative Prompt

```go
func generatePremiumImage() error {
    ctx := context.Background()
    client, err := generate.NewBedrockSDKClient(ctx, "us-west-2")
    if err != nil {
        return err
    }
    defer client.Close()

    options := models.GenerateOptions{
        Model:          "amazon.nova-canvas-v1:0",
        Size:           "1024x1792",
        Style:          "premium", // Triggers premium quality
        NegativePrompt: "blur, low quality, distorted, deformed",
        Seed:           42, // For reproducibility
    }

    img, err := client.GenerateImage(ctx,
        "a photorealistic portrait of a wise old wizard",
        options)
    if err != nil {
        return err
    }

    return os.WriteFile("wizard.png", img.Data, 0644)
}
```

### Example 3: With Timeout and Error Handling

```go
func generateWithTimeout() error {
    // Create context with 30-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    client, err := generate.NewBedrockSDKClient(ctx, "us-east-1")
    if err != nil {
        return fmt.Errorf("client creation failed: %w", err)
    }
    defer client.Close()

    options := models.GenerateOptions{
        Size: "2048x2048", // Max resolution
    }

    img, err := client.GenerateImage(ctx, "abstract art", options)
    if err != nil {
        // Check for specific errors
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("generation timed out after 30 seconds")
        }
        return fmt.Errorf("generation failed: %w", err)
    }

    return os.WriteFile("abstract.png", img.Data, 0644)
}
```

---

## Best Practices

### 1. Always Use Context with Timeout

```go
// ❌ Bad: No timeout
ctx := context.Background()

// ✅ Good: 60-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
```

### 2. Implement Circuit Breaker

Already implemented in the client using `gobreaker`. Prevents cascading failures.

### 3. Use Exponential Backoff for Retries

Already implemented. Max 3 retries with exponential backoff.

### 4. Validate Inputs Before Sending

```go
// Validate dimensions
if width < 512 || width > 2048 || height < 512 || height > 2048 {
    return fmt.Errorf("invalid dimensions: %dx%d (must be 512-2048)", width, height)
}

// Validate total pixels
totalPixels := width * height
if totalPixels < 262144 || totalPixels > 4194304 {
    return fmt.Errorf("invalid total pixels: %d (must be 262,144-4,194,304)", totalPixels)
}

// Validate seed
if seed < 0 || seed > 858993459 {
    return fmt.Errorf("invalid seed: %d (must be 0-858,993,459)", seed)
}
```

### 5. Log Verbose Information for Debugging

```go
// Enable verbose mode
export GIMAGE_VERBOSE=true

// Or in code
viper.Set("verbose", true)
```

### 6. Handle Credentials Securely

```go
// ✅ Good: Use environment variables or IAM roles
// No credentials in code

// ❌ Bad: Hardcoded credentials
// accessKey := "AKIAIOSFODNN7EXAMPLE" // NEVER DO THIS
```

### 7. Monitor API Usage and Costs

```go
// Log every generation for cost tracking
log.Printf("Generated image: model=%s, quality=%s, size=%dx%d, cost=$%.4f",
    modelID, quality, width, height, estimatedCost)
```

### 8. Use Appropriate Image Sizes

```go
// For web thumbnails
Size: "512x512"   // Fast, cheap ($0.04)

// For social media
Size: "1024x1024" // Good balance

// For print/high-res
Size: "2048x2048" // Premium quality ($0.08 if premium)
```

### 9. Implement Graceful Degradation

```go
// Try premium, fallback to standard
options := models.GenerateOptions{Style: "premium"}
img, err := client.GenerateImage(ctx, prompt, options)

if err != nil && contains(err.Error(), "quota") {
    // Fallback to standard quality
    options.Style = "standard"
    img, err = client.GenerateImage(ctx, prompt, options)
}
```

### 10. Cache Results When Possible

```go
// Use seed for reproducibility
cacheKey := fmt.Sprintf("%s:%d:%s", prompt, seed, size)
if cachedImage, ok := cache.Get(cacheKey); ok {
    return cachedImage
}

// Generate and cache
img, err := client.GenerateImage(ctx, prompt, options)
cache.Set(cacheKey, img)
```

---

## IAM Permissions Required

### Minimum Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:aws:bedrock:*::foundation-model/amazon.nova-canvas-v1:0"
      ]
    }
  ]
}
```

### Recommended Policy (Includes Model Access Request)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:ListFoundationModels",
        "bedrock:GetFoundationModel"
      ],
      "Resource": "*"
    }
  ]
}
```

### Using AWS Managed Policy

Simply attach `AmazonBedrockFullAccess` to your IAM user/role.

---

## Testing Strategy

### Unit Tests (Mock AWS SDK)

```go
// Use interfaces for mocking
type BedrockRuntimeAPI interface {
    InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput,
                opts ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

// In tests
type mockBedrockClient struct {
    invokeModelFunc func(ctx context.Context, input *bedrockruntime.InvokeModelInput) (*bedrockruntime.InvokeModelOutput, error)
}

func (m *mockBedrockClient) InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput, opts ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
    return m.invokeModelFunc(ctx, input)
}
```

### Integration Tests (Real API - Manual Only)

```go
// +build integration

func TestBedrockIntegration(t *testing.T) {
    if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
        t.Skip("AWS credentials not configured")
    }

    ctx := context.Background()
    client, err := generate.NewBedrockSDKClient(ctx, "us-east-1")
    require.NoError(t, err)

    // Test real generation (costs $0.04)
    img, err := client.GenerateImage(ctx, "test image", models.GenerateOptions{})
    require.NoError(t, err)
    require.NotEmpty(t, img.Data)
}
```

---

## Troubleshooting

### Issue: "AccessDeniedException"

**Cause**: IAM permissions missing
**Solution**:
1. Check IAM policy includes `bedrock:InvokeModel`
2. Verify model access granted in Bedrock console
3. Check region matches (us-east-1, us-west-2)

### Issue: "ResourceNotFoundException"

**Cause**: Model not available in region
**Solution**:
1. Use supported region (us-east-1 or us-west-2)
2. Verify model ID: `amazon.nova-canvas-v1:0`
3. Request model access in Bedrock console

### Issue: "ThrottlingException"

**Cause**: Rate limit exceeded (10 req/sec)
**Solution**:
1. Implement exponential backoff (already done)
2. Use circuit breaker (already done)
3. Reduce request rate

### Issue: "ValidationException: Invalid dimensions"

**Cause**: Image size out of bounds
**Solution**:
1. Use dimensions between 512-2048
2. Total pixels: 262,144 - 4,194,304
3. Check aspect ratio support

---

## References

- [AWS Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Nova Canvas User Guide](https://docs.aws.amazon.com/nova/latest/userguide/image-gen-req-resp-structure.html)
- [IAM Permissions for Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/security-iam.html)
- [Bedrock API Reference](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_InvokeModel.html)

---

## Summary

This guide provides production-ready code for integrating AWS Bedrock Nova Canvas with gimage. Key takeaways:

1. ✅ Use official AWS SDK for Go v2
2. ✅ Implement proper error handling for all AWS error types
3. ✅ Support multiple authentication methods
4. ✅ Use circuit breaker and retry logic
5. ✅ Validate inputs before sending to API
6. ✅ Log verbose information for debugging
7. ✅ Handle costs and quotas transparently
8. ✅ Test with mocks for unit tests, real API for integration tests

Ready to implement! Follow the AWS_NOVA_CANVAS_PLAN.md for step-by-step execution.
