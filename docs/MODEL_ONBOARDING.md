# Model Onboarding Guide

**Last Updated**: 2025-11-02

This guide provides a comprehensive process for adding new AI image generation models and providers to gimage. Use this checklist when integrating services like AWS Bedrock Nova Canvas, Azure AI, Anthropic, or any other image generation API.

---

## Table of Contents

1. [Overview](#overview)
2. [Phase 1: Research](#phase-1-research)
3. [Phase 2: Design](#phase-2-design)
4. [Phase 3: Implementation](#phase-3-implementation)
5. [Phase 4: Testing](#phase-4-testing)
6. [Phase 5: Documentation](#phase-5-documentation)
7. [Phase 6: Integration](#phase-6-integration)
8. [Checklist Summary](#checklist-summary)
9. [Example: AWS Bedrock Nova Canvas](#example-aws-bedrock-nova-canvas)

---

## Overview

### Architecture Principles

Gimage follows a **multi-backend architecture** where each provider implements a common interface:

```go
type ImageGenerator interface {
    GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error)
    Close() error
}
```

**Key Architectural Patterns:**
- **Client files**: `internal/generate/<provider>_rest.go` or `<provider>_sdk.go`
- **Model metadata**: Centralized in `internal/generate/models.go`
- **Configuration**: Markdown-based `~/.gimage/config.md`
- **Authentication**: Interactive `gimage auth <provider>` commands
- **Auto-detection**: Model name → API backend mapping
- **MCP integration**: Automatic exposure of all models via MCP server

---

## Phase 1: Research

### 1.1 API/SDK Documentation Review

**Goal**: Understand the provider's API thoroughly before writing any code.

#### Read Official Documentation
- [ ] **API Reference**: Find the REST API documentation
- [ ] **SDK Documentation**: Check if an official Go SDK exists
- [ ] **Authentication**: Understand auth mechanisms (API keys, OAuth, IAM, etc.)
- [ ] **Rate Limits**: Document requests per second/minute/day
- [ ] **Pricing**: Cost per image, free tier, batch discounts
- [ ] **Regions/Endpoints**: Available regions and endpoint URLs

#### Key Questions to Answer

**Authentication**:
- What credentials are required? (API key, access key/secret, service account, OAuth)
- How are credentials provided? (HTTP headers, query params, AWS signature)
- Are there multiple auth modes? (e.g., Vertex has API key + service account)
- What environment variables are standard? (e.g., `AWS_ACCESS_KEY_ID`)

**Request Format**:
- What is the API endpoint structure?
- What HTTP method? (POST, GET)
- What request body format? (JSON, multipart/form-data)
- Required vs. optional parameters?
- How is the prompt sent? (JSON field, form field)

**Response Format**:
- How is the image returned? (base64, URL, binary)
- What is the response structure? (JSON, headers)
- What metadata is included? (model version, generation ID, etc.)
- Error response format?

**Model Capabilities**:
- [ ] Supported image sizes (minimum, maximum, aspect ratios)
- [ ] Style controls (photorealistic, artistic, etc.)
- [ ] Negative prompts supported?
- [ ] Seed/reproducibility support?
- [ ] Maximum prompt length (characters or tokens)
- [ ] Supported output formats (PNG, JPG, WebP)

**Rate Limits & Quotas**:
- Requests per minute/hour/day?
- Concurrent request limits?
- Free tier limits?
- How are rate limits communicated? (HTTP headers, error codes)

**Error Handling**:
- Retryable error codes (429, 503, 500)
- Non-retryable errors (401, 403, 400)
- Circuit breaker scenarios
- Timeout recommendations

#### Research Artifacts

Create these documents during research:

**API_NOTES.md** (temporary, for your reference):
```markdown
# Provider: AWS Bedrock Nova Canvas

## Authentication
- Uses AWS Signature V4
- Credentials: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN (optional)
- Supports AWS profiles via AWS_PROFILE
- IAM permissions required: bedrock:InvokeModel

## Endpoint
- Format: https://bedrock-runtime.{region}.amazonaws.com/model/{model-id}/invoke
- Regions: us-east-1, us-west-2, eu-west-1
- Default region: us-east-1

## Request Format
POST /model/{model-id}/invoke
Headers:
  - Content-Type: application/json
  - AWS Signature V4 auth headers

Body:
{
  "textToImageParams": {
    "text": "prompt here"
  },
  "taskType": "TEXT_IMAGE",
  "imageGenerationConfig": {
    "numberOfImages": 1,
    "quality": "standard",
    "height": 1024,
    "width": 1024,
    "cfgScale": 7.0,
    "seed": 0
  }
}

## Response Format
{
  "images": ["base64-encoded-image-data"],
  "error": null
}

## Models
- amazon.nova-canvas-v1:0 (latest)
- Max size: 2048x2048
- Aspect ratios: 1:1, 16:9, 9:16, 4:3, 3:4

## Pricing
- $0.04 per standard quality image
- $0.08 per premium quality image

## Rate Limits
- 10 requests per second per account
- 1000 requests per hour
```

---

## Phase 2: Design

### 2.1 Architecture Decisions

#### Client Implementation Strategy

**Decision Matrix**:

| Factor | REST Client | Official SDK |
|--------|------------|--------------|
| SDK Quality | - | Go SDK exists, well-maintained |
| Dependencies | No external deps | May add large dependencies |
| Flexibility | Full control | Constrained by SDK design |
| Maintenance | Manual updates | SDK updates automatically |
| Auth Complexity | Implement manually | SDK handles auth |

**Recommendation**:
- **Use SDK if**: Official Go SDK exists, handles complex auth (e.g., AWS Signature V4), and is well-maintained
- **Use REST if**: Simple API key auth, no SDK, or SDK is heavyweight/unmaintained

For **AWS Bedrock Nova Canvas**: Use SDK (AWS SDK v2) because AWS Signature V4 is complex and the SDK is excellent.

#### File Structure

```
internal/generate/
├── bedrock_sdk.go           # Client implementation
├── bedrock_sdk_test.go      # Unit tests
├── models.go                # Add model metadata
```

### 2.2 Model Metadata Design

Plan the `ModelInfo` entry for `models.go`:

```go
{
    Name:        "amazon.nova-canvas-v1:0",
    DisplayName: "AWS Nova Canvas",
    API:         "bedrock",  // New API identifier
    Quality:     "high",
    Description: "Amazon's Nova Canvas model via AWS Bedrock",
    Priority:    7,  // Adjust based on cost/quality
    RequiresAuth: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},

    Pricing: PricingInfo{
        CostPerImage:       float64Ptr(0.04),  // Standard quality
        CostPerImageHD:     float64Ptr(0.08),  // Premium quality
        FreeTier:           false,
        MaxResolution:      "2048x2048",
        SupportedSizes:     []string{"512x512", "1024x1024", "1024x1792", "1792x1024", "2048x2048"},
        RateLimits: RateLimits{
            RequestsPerMinute: intPtr(600),  // 10 req/sec * 60
        },
        BillingUnit: "per_image",
        Currency:    "USD",
        PricingTier: "standard",
        LastUpdated: "2025-11-02",
    },

    Capabilities: ModelCapabilities{
        SupportsStyles:         true,
        SupportsNegativePrompt: true,
        SupportsSeed:           true,
        SupportedStyles:        []string{"photorealistic", "artistic"},
        MaxPromptLength:        1024,  // Check AWS docs
    },
}
```

### 2.3 Configuration Schema

Add to `~/.gimage/config.md`:

```markdown
# AWS Bedrock Configuration
**aws_access_key_id**: AKIAIOSFODNN7EXAMPLE
**aws_secret_access_key**: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
**aws_region**: us-east-1
**aws_profile**: default  # Optional, alternative to keys
```

Add to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields ...

    // AWS Bedrock
    AWSAccessKeyID     string
    AWSSecretAccessKey string
    AWSRegion          string
    AWSProfile         string
}
```

---

## Phase 3: Implementation

### 3.1 Client Implementation

**File**: `internal/generate/bedrock_sdk.go`

#### Step-by-Step Implementation

**Step 1: Package and Imports**

```go
package generate

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/apresai/gimage/pkg/models"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
    "github.com/sony/gobreaker"
    "github.com/spf13/viper"
)
```

**Step 2: Client Structure**

```go
// BedrockSDKClient uses AWS Bedrock API for image generation
type BedrockSDKClient struct {
    client         *bedrockruntime.Client
    region         string
    verbose        bool
    circuitBreaker *gobreaker.CircuitBreaker
}
```

**Step 3: Constructor**

```go
// NewBedrockSDKClient creates a new AWS Bedrock SDK client
func NewBedrockSDKClient(ctx context.Context, region string) (*BedrockSDKClient, error) {
    if region == "" {
        region = os.Getenv("AWS_REGION")
        if region == "" {
            region = "us-east-1" // Default region
        }
    }

    // Load AWS config (uses environment variables or AWS config files)
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion(region),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    // Create Bedrock Runtime client
    client := bedrockruntime.NewFromConfig(cfg)

    // Check if verbose mode is enabled
    verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true"

    return &BedrockSDKClient{
        client:         client,
        region:         region,
        verbose:        verbose,
        circuitBreaker: newCircuitBreaker("BedrockAPI"),
    }, nil
}
```

**Step 4: Verbose Logging**

```go
func (c *BedrockSDKClient) logVerbose(format string, args ...interface{}) {
    if c.verbose {
        fmt.Fprintf(os.Stderr, "[BEDROCK-SDK] "+format+"\n", args...)
    }
}
```

**Step 5: GenerateImage Implementation**

```go
func (c *BedrockSDKClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    // Validate prompt
    if err := ValidatePrompt(prompt); err != nil {
        return nil, err
    }

    // Enhance prompt for better results
    enhancedPrompt := EnhancePrompt(prompt)

    // Use custom model if provided
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

        // Don't sleep after the last attempt
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

func (c *BedrockSDKClient) generateWithRetry(ctx context.Context, modelID, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    // Parse dimensions
    width, height := parseDimensions(options.Size)

    c.logVerbose("Building request for model: %s", modelID)
    c.logVerbose("Requested dimensions: %dx%d", width, height)
    c.logVerbose("Prompt: %s", prompt)

    // Build request payload according to Nova Canvas API
    // NOTE: Check AWS documentation for exact request format
    requestBody := map[string]interface{}{
        "textToImageParams": map[string]interface{}{
            "text": prompt,
        },
        "taskType": "TEXT_IMAGE",
        "imageGenerationConfig": map[string]interface{}{
            "numberOfImages": 1,
            "quality":        "standard",
            "height":         height,
            "width":          width,
            "cfgScale":       7.0,
        },
    }

    // Add negative prompt if provided
    if options.NegativePrompt != "" {
        requestBody["textToImageParams"].(map[string]interface{})["negativeText"] = options.NegativePrompt
    }

    // Add seed if provided
    if options.Seed != 0 {
        requestBody["imageGenerationConfig"].(map[string]interface{})["seed"] = options.Seed
    }

    // Marshal to JSON
    requestJSON, err := json.Marshal(requestBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    c.logVerbose("Request body: %s", string(requestJSON))

    // Invoke model
    c.logVerbose("Invoking model: %s", modelID)

    input := &bedrockruntime.InvokeModelInput{
        ModelId:     aws.String(modelID),
        ContentType: aws.String("application/json"),
        Accept:      aws.String("application/json"),
        Body:        requestJSON,
    }

    result, err := c.client.InvokeModel(ctx, input)
    if err != nil {
        c.logVerbose("InvokeModel failed: %v", err)
        return nil, c.handleSDKError(err)
    }

    // Parse response
    var response struct {
        Images []string `json:"images"` // Base64-encoded images
        Error  string   `json:"error"`
    }

    if err := json.Unmarshal(result.Body, &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    if response.Error != "" {
        return nil, fmt.Errorf("API error: %s", response.Error)
    }

    if len(response.Images) == 0 {
        return nil, fmt.Errorf("no image generated")
    }

    // Decode base64 image
    imageData, err := base64.StdEncoding.DecodeString(response.Images[0])
    if err != nil {
        return nil, fmt.Errorf("failed to decode base64 image: %w", err)
    }

    c.logVerbose("Successfully generated image: %d bytes", len(imageData))

    return &models.GeneratedImage{
        Data:   imageData,
        Format: "png",  // Check AWS docs for actual format
        Width:  width,
        Height: height,
        Metadata: map[string]string{
            "model":  modelID,
            "prompt": prompt,
            "style":  options.Style,
            "api":    "bedrock",
            "region": c.region,
        },
    }, nil
}
```

**Step 6: Error Handling**

```go
func (c *BedrockSDKClient) handleSDKError(err error) error {
    errStr := err.Error()

    // Check for specific AWS error types
    if contains(errStr, "AccessDeniedException") {
        return fmt.Errorf("permission denied: check IAM permissions for bedrock:InvokeModel")
    }

    if contains(errStr, "ResourceNotFoundException") {
        return fmt.Errorf("model not found: verify model ID and region")
    }

    if contains(errStr, "ThrottlingException") {
        return fmt.Errorf("rate limit exceeded: too many requests")
    }

    if contains(errStr, "ServiceUnavailableException") {
        return fmt.Errorf("service unavailable: try again later")
    }

    return fmt.Errorf("AWS Bedrock error: %w", err)
}
```

**Step 7: Close Method**

```go
func (c *BedrockSDKClient) Close() error {
    // AWS SDK clients don't require explicit closing
    return nil
}
```

### 3.2 Configuration Updates

**File**: `internal/config/config.go`

```go
// Add to Config struct
type Config struct {
    // ... existing fields ...

    // AWS Bedrock
    AWSAccessKeyID     string
    AWSSecretAccessKey string
    AWSRegion          string
    AWSProfile         string
}

// Add to parseMarkdownConfig
switch key {
    // ... existing cases ...
    case "aws_access_key_id":
        cfg.AWSAccessKeyID = value
    case "aws_secret_access_key":
        cfg.AWSSecretAccessKey = value
    case "aws_region":
        cfg.AWSRegion = value
    case "aws_profile":
        cfg.AWSProfile = value
}

// Add to SaveConfig
if cfg.AWSAccessKeyID != "" {
    content.WriteString(fmt.Sprintf("**aws_access_key_id**: %s\n", cfg.AWSAccessKeyID))
}
if cfg.AWSSecretAccessKey != "" {
    content.WriteString(fmt.Sprintf("**aws_secret_access_key**: %s\n", cfg.AWSSecretAccessKey))
}
if cfg.AWSRegion != "" {
    content.WriteString(fmt.Sprintf("**aws_region**: %s\n", cfg.AWSRegion))
}
if cfg.AWSProfile != "" {
    content.WriteString(fmt.Sprintf("**aws_profile**: %s\n", cfg.AWSProfile))
}

// Add helper functions
func GetAWSRegion() string {
    // Priority: environment variable > config file > default
    if region := os.Getenv("AWS_REGION"); region != "" {
        return region
    }

    cfg, err := LoadConfig()
    if err == nil && cfg.AWSRegion != "" {
        return cfg.AWSRegion
    }

    return "us-east-1"
}

func HasBedrockCredentials() bool {
    // Check for AWS credentials (access keys or profile)
    if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
        return true
    }

    if os.Getenv("AWS_PROFILE") != "" {
        return true
    }

    cfg, err := LoadConfig()
    if err != nil {
        return false
    }

    if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
        return true
    }

    if cfg.AWSProfile != "" {
        return true
    }

    return false
}
```

### 3.3 Model Registry Updates

**File**: `internal/generate/models.go`

```go
// Add to constants
const (
    // ... existing models ...

    // AWS Bedrock models
    ModelNovaCanvasV1 = "amazon.nova-canvas-v1:0"
)

// Add to ModelAliases
var ModelAliases = map[string]string{
    // ... existing aliases ...
    "nova-canvas":     "amazon.nova-canvas-v1:0",
    "nova":            "amazon.nova-canvas-v1:0",
    "bedrock-canvas":  "amazon.nova-canvas-v1:0",
}

// Add to AvailableModels()
func AvailableModels() []ModelInfo {
    return []ModelInfo{
        // ... existing models ...

        {
            Name:        ModelNovaCanvasV1,
            DisplayName: "AWS Nova Canvas",
            API:         "bedrock",
            Quality:     "high",
            Description: "Amazon's Nova Canvas model via AWS Bedrock",
            Priority:    7,
            RequiresAuth: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
            MaxSize:      "2048x2048",
            Free:         false,

            Pricing: PricingInfo{
                CostPerImage:       float64Ptr(0.04),
                CostPerImageHD:     float64Ptr(0.08),
                FreeTier:           false,
                MaxResolution:      "2048x2048",
                SupportedSizes:     []string{"512x512", "1024x1024", "1024x1792", "1792x1024", "2048x2048"},
                RateLimits: RateLimits{
                    RequestsPerMinute: intPtr(600),  // 10/sec * 60
                },
                BillingUnit: "per_image",
                Currency:    "USD",
                PricingTier: "standard",
                LastUpdated: "2025-11-02",
            },

            Capabilities: ModelCapabilities{
                SupportsStyles:         true,
                SupportsNegativePrompt: true,
                SupportsSeed:           true,
                SupportedStyles:        []string{"photorealistic", "artistic"},
                MaxPromptLength:        1024,
            },
        },
    }
}

// Update ValidateConfig
func ValidateConfig(cfg *Config) error {
    // ... existing validation ...

    // Add bedrock to valid APIs
    validAPIs := map[string]bool{
        "gemini":  true,
        "vertex":  true,
        "bedrock": true,  // Add this
    }
}
```

### 3.4 CLI Integration

**File**: `internal/cli/generate.go`

Update the generation logic to handle the new API:

```go
// In runGenerate function, add bedrock case
} else if selectedAPI == "bedrock" {
    // Use AWS Bedrock
    region := os.Getenv("AWS_REGION")
    if region == "" {
        region = config.GetAWSRegion()
    }

    modelName := model
    if modelName == "" {
        modelName = "amazon.nova-canvas-v1:0"
    }

    // Get model info and announce selection
    modelInfo, _ := generate.GetModelInfo(modelName)
    if modelInfo != nil {
        printInfo("Using: %s (%s API)", modelInfo.DisplayName, modelInfo.API)
        printInfo("Pricing: %s", generate.FormatPricingDisplay(modelInfo))

        // Calculate estimated cost
        cost, _, explanation := generate.GetEstimatedCost(modelInfo, size, 1)
        printVerbose("Estimated: %s", explanation)

        if cost > 0.05 {
            fmt.Fprintf(os.Stderr, "⚠️  %s costs $%.4f/image\n", modelInfo.DisplayName, *modelInfo.Pricing.CostPerImage)
        }
    }

    printInfo("Generating image...")

    client, err := generate.NewBedrockSDKClient(ctx, region)
    if err != nil {
        return fmt.Errorf("failed to create Bedrock client: %w", err)
    }
    defer client.Close()

    generatedImage, err = client.GenerateImage(ctx, prompt, options)
} else {
    return fmt.Errorf("invalid API: %s (must be 'gemini', 'vertex', or 'bedrock')", selectedAPI)
}
```

Update auto-detection logic:

```go
// In auto-detection section
hasGemini := config.HasGeminiCredentials()
hasVertex := config.HasVertexCredentials()
hasBedrock := config.HasBedrockCredentials()

if hasGemini && !hasVertex && !hasBedrock {
    selectedAPI = "gemini"
} else if hasVertex && !hasGemini && !hasBedrock {
    selectedAPI = "vertex"
} else if hasBedrock && !hasGemini && !hasVertex {
    selectedAPI = "bedrock"
    printVerbose("Auto-detected Bedrock API (found credentials)")
} else if ... {
    // Handle multiple credentials case
}
```

### 3.5 Authentication Command

**File**: `internal/cli/auth.go`

```go
// Add auth command
var authBedrockCmd = &cobra.Command{
    Use:   "bedrock",
    Short: "Configure AWS Bedrock authentication",
    Long: `Interactive setup for AWS Bedrock credentials.

AWS Bedrock supports two authentication modes:
  1. Access Keys - AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
  2. AWS Profile - Uses credentials from ~/.aws/credentials`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return setupBedrockAuth()
    },
}

func init() {
    authCmd.AddCommand(authBedrockCmd)
}

func setupBedrockAuth() error {
    reader := bufio.NewReader(os.Stdin)

    // Load existing config
    existingCfg, err := config.LoadConfig()
    if err != nil {
        existingCfg = &config.Config{
            DefaultAPI:     "gemini",
            DefaultModel:   "gemini-2.5-flash-image",
            DefaultSize:    "1024x1024",
            VertexLocation: "us-central1",
            LogLevel:       "info",
        }
    }

    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Println("  AWS Bedrock Authentication Setup")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Println()
    fmt.Println("Choose your authentication mode:")
    fmt.Println()
    fmt.Println("  1. Access Keys - Use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
    fmt.Println("     • Best for: Testing, development")
    fmt.Println("     • Get from: AWS Console > IAM > Users > Security Credentials")
    fmt.Println()
    fmt.Println("  2. AWS Profile - Use credentials from ~/.aws/credentials")
    fmt.Println("     • Best for: Local development with AWS CLI configured")
    fmt.Println("     • Requires: aws configure command already run")
    fmt.Println()

    mode := promptWithDefault(reader, "Choose mode (1 or 2)", "2", false)

    fmt.Println()

    if mode == "1" {
        return setupBedrockAccessKeys(reader, existingCfg)
    } else {
        return setupBedrockProfile(reader, existingCfg)
    }
}

func setupBedrockAccessKeys(reader *bufio.Reader, existingCfg *config.Config) error {
    fmt.Println("━━━ Access Keys Setup ━━━")
    fmt.Println()

    // Access Key ID
    accessKeyID := promptWithDefault(reader, "AWS Access Key ID", existingCfg.AWSAccessKeyID, true)

    // Secret Access Key
    secretAccessKey := promptWithDefault(reader, "AWS Secret Access Key", existingCfg.AWSSecretAccessKey, true)

    // Region
    region := promptWithDefault(reader, "AWS Region", existingCfg.AWSRegion, false)
    if region == "" {
        region = "us-east-1"
    }

    // Update config
    existingCfg.AWSAccessKeyID = accessKeyID
    existingCfg.AWSSecretAccessKey = secretAccessKey
    existingCfg.AWSRegion = region
    existingCfg.AWSProfile = "" // Clear profile

    // Save config
    if err := config.SaveConfig(existingCfg); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    configPath := config.GetConfigPath()
    fmt.Println()
    fmt.Println("✓ AWS Bedrock configured successfully!")
    fmt.Printf("  Location: %s\n", configPath)
    fmt.Println()
    fmt.Println("You can now use AWS Bedrock with:")
    fmt.Println("  gimage generate --api bedrock \"your prompt here\"")
    fmt.Println()

    return nil
}

func setupBedrockProfile(reader *bufio.Reader, existingCfg *config.Config) error {
    fmt.Println("━━━ AWS Profile Setup ━━━")
    fmt.Println()

    // Profile name
    profile := promptWithDefault(reader, "AWS Profile name", existingCfg.AWSProfile, false)
    if profile == "" {
        profile = "default"
    }

    // Region
    region := promptWithDefault(reader, "AWS Region", existingCfg.AWSRegion, false)
    if region == "" {
        region = "us-east-1"
    }

    fmt.Println()
    fmt.Println("Using AWS Profile authentication.")
    fmt.Println()
    fmt.Println("Make sure your profile is configured in ~/.aws/credentials")
    fmt.Println()

    // Update config
    existingCfg.AWSProfile = profile
    existingCfg.AWSRegion = region
    existingCfg.AWSAccessKeyID = ""     // Clear access keys
    existingCfg.AWSSecretAccessKey = "" // Clear secret key

    // Save config
    if err := config.SaveConfig(existingCfg); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    configPath := config.GetConfigPath()
    fmt.Println()
    fmt.Println("✓ AWS Bedrock configured successfully!")
    fmt.Printf("  Location: %s\n", configPath)
    fmt.Println()
    fmt.Println("You can now use AWS Bedrock with:")
    fmt.Println("  gimage generate --api bedrock \"your prompt here\"")
    fmt.Println()

    return nil
}
```

---

## Phase 4: Testing

### 4.1 Unit Tests

**CRITICAL**: Do NOT mock cloud provider APIs. Test logic, not mocked behavior.

**File**: `internal/generate/bedrock_test.go`

```go
package generate

import (
    "testing"

    "github.com/apresai/gimage/pkg/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// ✅ GOOD: Test request building logic
func TestBuildNovaCanvasRequest(t *testing.T) {
    tests := []struct {
        name    string
        prompt  string
        options models.GenerateOptions
        want    NovaCanvasRequest
    }{
        {
            name:   "basic request",
            prompt: "a sunset over mountains",
            options: models.GenerateOptions{
                Size: "1024x1024",
            },
            want: NovaCanvasRequest{
                TaskType: "TEXT_IMAGE",
                TextToImageParams: NovaCanvasTextToImageParams{
                    Text: "a sunset over mountains",
                },
                ImageGenerationConfig: NovaCanvasImageConfig{
                    NumberOfImages: 1,
                    Quality:        "standard",
                    Height:         1024,
                    Width:          1024,
                    CfgScale:       7.0,
                },
            },
        },
        {
            name:   "with negative prompt and seed",
            prompt: "portrait",
            options: models.GenerateOptions{
                Size:           "512x512",
                NegativePrompt: "blur, distorted",
                Seed:           42,
            },
            want: NovaCanvasRequest{
                TaskType: "TEXT_IMAGE",
                TextToImageParams: NovaCanvasTextToImageParams{
                    Text:         "portrait",
                    NegativeText: "blur, distorted",
                },
                ImageGenerationConfig: NovaCanvasImageConfig{
                    NumberOfImages: 1,
                    Quality:        "standard",
                    Height:         512,
                    Width:          512,
                    CfgScale:       7.0,
                    Seed:           42,
                },
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := buildNovaCanvasRequest(tt.prompt, tt.options)
            assert.Equal(t, tt.want.TaskType, got.TaskType)
            assert.Equal(t, tt.want.TextToImageParams.Text, got.TextToImageParams.Text)
            assert.Equal(t, tt.want.ImageGenerationConfig.Width, got.ImageGenerationConfig.Width)
            assert.Equal(t, tt.want.ImageGenerationConfig.Seed, got.ImageGenerationConfig.Seed)
        })
    }
}

// ✅ GOOD: Test response parsing with real examples from AWS docs
func TestParseNovaCanvasResponse(t *testing.T) {
    tests := []struct {
        name    string
        body    []byte
        wantErr bool
    }{
        {
            name: "successful response",
            body: []byte(`{"images": ["iVBORw0KGgoAAAANSUhEUgAAAAUA..."], "error": null}`),
            wantErr: false,
        },
        {
            name: "error response",
            body: []byte(`{"images": [], "error": "ValidationException: Invalid dimensions"}`),
            wantErr: true,
        },
        {
            name: "empty images array",
            body: []byte(`{"images": []}`),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := parseNovaCanvasResponse(tt.body)

            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.NotEmpty(t, result.Images)
            }
        })
    }
}

// ✅ GOOD: Test error handling logic
func TestHandleSDKError(t *testing.T) {
    tests := []struct {
        name         string
        errStr       string
        wantContains string
    }{
        {
            name:         "access denied",
            errStr:       "AccessDeniedException: User not authorized",
            wantContains: "permission denied",
        },
        {
            name:         "throttling",
            errStr:       "ThrottlingException: Rate exceeded",
            wantContains: "rate limit exceeded",
        },
        {
            name:         "service unavailable",
            errStr:       "ServiceUnavailableException: Try again",
            wantContains: "service unavailable",
        },
        {
            name:         "resource not found",
            errStr:       "ResourceNotFoundException: Model not found",
            wantContains: "model not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := &BedrockSDKClient{region: "us-east-1"}
            err := client.handleSDKError(fmt.Errorf(tt.errStr))

            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantContains)
        })
    }
}

// ✅ GOOD: Test input validation
func TestValidateNovaCanvasInput(t *testing.T) {
    tests := []struct {
        name    string
        width   int
        height  int
        seed    int
        wantErr bool
    }{
        {
            name:    "valid dimensions",
            width:   1024,
            height:  1024,
            seed:    42,
            wantErr: false,
        },
        {
            name:    "width too small",
            width:   256,
            height:  1024,
            wantErr: true,
        },
        {
            name:    "seed out of range",
            width:   1024,
            height:  1024,
            seed:    999999999,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateNovaCanvasInput(tt.width, tt.height, tt.seed)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// ❌ DO NOT DO THIS - Mocking AWS SDK is worthless
// type mockBedrockClient struct { ... }
// func TestWithMockedAWS(t *testing.T) { ... }
```

### 4.2 Integration Tests (Manual Only - Real API Calls)

**IMPORTANT**: These tests make REAL API calls and cost REAL money. Run manually only.

**File**: `test/integration/bedrock_test.go`

```go
// +build integration

package integration

import (
    "context"
    "os"
    "testing"

    "github.com/apresai/gimage/internal/generate"
    "github.com/apresai/gimage/pkg/models"
    "github.com/stretchr/testify/require"
)

// IMPORTANT: These tests make REAL API calls to AWS Bedrock
// Cost: ~$0.04 per test run
// Run with: go test -tags=integration ./test/integration/...

func TestBedrockRealAPI(t *testing.T) {
    // Skip if no credentials
    if !hasAWSCredentials() {
        t.Skip("AWS credentials not configured (set AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY)")
    }

    t.Log("⚠️  This test will make a REAL API call and cost $0.04")

    ctx := context.Background()
    client, err := generate.NewBedrockSDKClient(ctx, "us-east-1")
    require.NoError(t, err)
    defer client.Close()

    options := models.GenerateOptions{
        Model: "amazon.nova-canvas-v1:0",
        Size:  "512x512", // Smaller size, but same cost
    }

    // Real API call (costs $0.04)
    img, err := client.GenerateImage(ctx, "simple test image", options)
    require.NoError(t, err)
    require.NotNil(t, img)
    require.NotEmpty(t, img.Data)
    require.Equal(t, "png", img.Format)
    require.Equal(t, 512, img.Width)
    require.Equal(t, 512, img.Height)

    t.Logf("✅ Successfully generated image: %d bytes", len(img.Data))
}

func TestBedrockErrorHandling(t *testing.T) {
    if !hasAWSCredentials() {
        t.Skip("AWS credentials not configured")
    }

    t.Log("⚠️  This test will make REAL API calls")

    ctx := context.Background()
    client, err := generate.NewBedrockSDKClient(ctx, "us-east-1")
    require.NoError(t, err)
    defer client.Close()

    // Test invalid dimensions (should fail without costing money)
    options := models.GenerateOptions{
        Size: "99999x99999", // Invalid size
    }

    _, err = client.GenerateImage(ctx, "test", options)
    require.Error(t, err)
    require.Contains(t, err.Error(), "validation")
}

func hasAWSCredentials() bool {
    return (os.Getenv("AWS_ACCESS_KEY_ID") != "" &&
            os.Getenv("AWS_SECRET_ACCESS_KEY") != "") ||
           os.Getenv("AWS_PROFILE") != ""
}
```

**Running Integration Tests**:

```bash
# Set credentials
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"

# Run integration tests (costs ~$0.04-0.08)
go test -tags=integration -v ./test/integration/bedrock_test.go

# Expected output:
# === RUN   TestBedrockRealAPI
# ⚠️  This test will make a REAL API call and cost $0.04
# ✅ Successfully generated image: 45123 bytes
# --- PASS: TestBedrockRealAPI (8.23s)
```

### 4.3 MCP Tool Testing

Test that the MCP server correctly exposes the new model:

```bash
# Start MCP server
gimage serve

# In another terminal, test with MCP client
# The generate_image tool should now accept bedrock models
```

**Test cases**:
1. List models - should include Nova Canvas
2. Generate with `--model amazon.nova-canvas-v1:0`
3. Auto-detection when only Bedrock credentials exist
4. Error handling for missing credentials

---

## Phase 5: Documentation

### 5.1 Update CLAUDE.md

Add Bedrock to the API Integration section:

```markdown
### AWS Bedrock Backend

**Implementation**: SDK client (`bedrock_sdk.go`)

**Setup**:
```bash
gimage auth bedrock
```

**Models**:
- `amazon.nova-canvas-v1:0` (latest Nova Canvas)

**Authentication Options**:
1. Access Keys via `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
2. AWS Profile via `AWS_PROFILE`

**Usage in Code**:
```go
region := "us-east-1"
client, err := generate.NewBedrockSDKClient(ctx, region)
defer client.Close()

img, err := client.GenerateImage(ctx, prompt, options)
```

**Pricing**:
- Standard quality: $0.04/image
- Premium quality: $0.08/image
- No free tier

**Rate Limits**:
- 10 requests per second
- 600 requests per minute
```

### 5.2 Update MCP_TOOLS.md

Add to supported models:

```markdown
### Supported Models

- **gemini-2.5-flash-image** (default, recommended)
- **gemini-2.0-flash-preview-image-generation**
- **imagen-3.0-generate-002** (requires Vertex AI)
- **imagen-4** (requires Vertex AI, highest quality)
- **amazon.nova-canvas-v1:0** (requires AWS Bedrock)
```

### 5.3 Update README.md

```markdown
## Supported AI Models

| Provider | Model | Max Resolution | Free Tier | Cost |
|----------|-------|----------------|-----------|------|
| Google Gemini | gemini-2.5-flash-image | 1792x1024 | ✅ 500/day | $0.039/image |
| Google Vertex | imagen-4 | 2048x2048 | ❌ | $0.04/image |
| AWS Bedrock | amazon.nova-canvas-v1:0 | 2048x2048 | ❌ | $0.04/image |

### AWS Bedrock Setup

1. **Install AWS CLI** (if not already installed):
   ```bash
   curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
   sudo installer -pkg AWSCLIV2.pkg -target /
   ```

2. **Configure credentials**:
   ```bash
   gimage auth bedrock
   ```

3. **Generate an image**:
   ```bash
   gimage generate --api bedrock "a sunset over mountains"
   ```
```

### 5.4 Create Migration Guide

**File**: `docs/BEDROCK_MIGRATION.md`

```markdown
# AWS Bedrock Migration Guide

This guide helps you migrate to AWS Bedrock or add it alongside existing providers.

## Prerequisites

- AWS Account with Bedrock access
- IAM permissions: `bedrock:InvokeModel`
- AWS CLI configured (optional, for profile auth)

## Step 1: Enable Bedrock in AWS Console

1. Go to AWS Console → Amazon Bedrock
2. Navigate to "Model access"
3. Request access to "Amazon Nova Canvas"
4. Wait for approval (usually instant)

## Step 2: Create IAM User (Access Key Method)

1. Go to IAM → Users → Create User
2. Attach policy: `AmazonBedrockFullAccess`
3. Create access key
4. Save Access Key ID and Secret Access Key

## Step 3: Configure gimage

```bash
gimage auth bedrock
```

Choose "Access Keys" and enter your credentials.

## Step 4: Test

```bash
gimage generate --api bedrock "test image"
```

## Troubleshooting

### "AccessDeniedException"
- Check IAM permissions
- Verify model access is granted in Bedrock console

### "ResourceNotFoundException"
- Check region (Nova Canvas availability varies)
- Verify model ID: `amazon.nova-canvas-v1:0`

### "ThrottlingException"
- Reduce request rate
- Check account limits in AWS Console
```

---

## Phase 6: Integration

### 6.1 go.mod Updates

Add AWS SDK dependency:

```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime
go mod tidy
```

### 6.2 Build & Test

```bash
# Run unit tests
make test

# Run linter
make lint

# Build binary
make build

# Manual test
./bin/gimage auth bedrock
./bin/gimage generate --api bedrock "test image" --verbose

# Test MCP server
./bin/gimage serve
```

### 6.3 Update Version

If this is a significant addition, bump version:

**File**: `cmd/gimage/main.go`

```go
const version = "0.3.0"  // Minor version bump for new provider
```

### 6.4 Update CHANGELOG.md

```bash
# Get current date
date +%Y-%m-%d
```

```markdown
## [0.3.0] - 2025-11-02

### Added
- AWS Bedrock Nova Canvas integration
- New `gimage auth bedrock` command for AWS credentials setup
- Support for `amazon.nova-canvas-v1:0` model
- AWS region configuration support
- Documentation for Bedrock setup and usage

### Changed
- Updated model auto-detection to include Bedrock
- Extended `--list-models` to show AWS models

### Developer Notes
- Added `internal/generate/bedrock_sdk.go` client
- Integrated AWS SDK v2 for Bedrock Runtime
- Added model metadata for Nova Canvas
```

---

## Checklist Summary

Use this checklist when onboarding a new model/provider:

### Research Phase
- [ ] Read API/SDK documentation thoroughly
- [ ] Document authentication mechanisms
- [ ] Identify request/response formats
- [ ] Document rate limits and pricing
- [ ] List model capabilities (sizes, styles, etc.)
- [ ] Document error codes and retry logic

### Design Phase
- [ ] Decide: REST client vs. SDK client
- [ ] Design model metadata structure
- [ ] Plan configuration schema
- [ ] Design authentication flow
- [ ] Plan CLI integration points

### Implementation Phase
- [ ] Create client file (`<provider>_sdk.go` or `<provider>_rest.go`)
- [ ] Implement `GenerateImage()` method
- [ ] Implement `Close()` method
- [ ] Add error handling and retries
- [ ] Update `models.go` with model metadata
- [ ] Update `config.go` with new config fields
- [ ] Add validation for new API type
- [ ] Create `gimage auth <provider>` command
- [ ] Integrate into `gimage generate` CLI
- [ ] Update auto-detection logic

### Testing Phase
- [ ] Write unit tests with mocking
- [ ] Create integration tests (manual)
- [ ] Test MCP server integration
- [ ] Test authentication flow
- [ ] Test error handling
- [ ] Test rate limiting/retries
- [ ] Test with various image sizes

### Documentation Phase
- [ ] Update CLAUDE.md with new backend section
- [ ] Update MCP_TOOLS.md supported models
- [ ] Update README.md with provider info
- [ ] Create migration guide if needed
- [ ] Update CHANGELOG.md

### Integration Phase
- [ ] Update go.mod with new dependencies
- [ ] Run `go mod tidy`
- [ ] Build and test locally
- [ ] Update version number
- [ ] Run full test suite
- [ ] Test MCP server end-to-end

### Release Phase
- [ ] Create feature branch
- [ ] Commit changes with clear messages
- [ ] Test on clean machine if possible
- [ ] Create PR with detailed description
- [ ] Tag release (if applicable)

---

## Example: AWS Bedrock Nova Canvas

Here's a concrete walkthrough for adding AWS Bedrock Nova Canvas:

### Quick Reference

**Provider**: AWS Bedrock
**Model**: amazon.nova-canvas-v1:0
**Client Type**: SDK (AWS SDK v2)
**Auth**: AWS Access Keys or AWS Profile
**Pricing**: $0.04/image (standard), $0.08/image (premium)
**Max Resolution**: 2048x2048
**Rate Limit**: 10 req/sec, 600 req/min

### File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `internal/generate/bedrock_sdk.go` | Create | Client implementation |
| `internal/generate/bedrock_sdk_test.go` | Create | Unit tests |
| `internal/generate/models.go` | Modify | Add Nova Canvas metadata |
| `internal/config/config.go` | Modify | Add AWS config fields |
| `internal/cli/auth.go` | Modify | Add `auth bedrock` command |
| `internal/cli/generate.go` | Modify | Add bedrock API case |
| `go.mod` | Modify | Add AWS SDK dependencies |
| `docs/CLAUDE.md` | Modify | Add Bedrock section |
| `docs/MCP_TOOLS.md` | Modify | Add Nova Canvas to models |
| `docs/BEDROCK_MIGRATION.md` | Create | Setup guide |
| `CHANGELOG.md` | Modify | Document new feature |

### Time Estimate

- **Research**: 2-3 hours (reading AWS docs, testing API)
- **Design**: 1 hour (planning architecture)
- **Implementation**: 4-6 hours (client, config, CLI, auth)
- **Testing**: 2-3 hours (unit tests, integration tests, manual testing)
- **Documentation**: 1-2 hours (updating docs, writing guides)

**Total**: ~10-15 hours for a complete, production-ready integration

### Dependencies

```bash
# AWS SDK v2
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime@latest
go get github.com/aws/aws-sdk-go-v2/credentials@latest
```

### Testing Commands

```bash
# Unit tests
go test ./internal/generate/bedrock_sdk_test.go -v

# Integration tests (requires credentials)
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"
go test -tags=integration ./test/integration/bedrock_test.go -v

# Manual CLI test
gimage auth bedrock
gimage generate --api bedrock "a sunset over mountains" --verbose
gimage generate --model amazon.nova-canvas-v1:0 "test image" -o test.png
gimage generate --list-models | grep -i bedrock
```

---

## Best Practices

### Code Quality
1. **Follow existing patterns**: Match the style of `gemini_rest.go` and `vertex_rest.go`
2. **Error messages**: Provide actionable error messages with suggestions
3. **Logging**: Use `logVerbose()` for debugging without cluttering output
4. **Comments**: Document complex logic, especially API-specific quirks

### Testing
1. **Mock API calls**: Never make real API calls in unit tests (they cost money)
2. **Test error paths**: Test authentication failures, rate limits, timeouts
3. **Integration tests**: Mark with `// +build integration` build tag
4. **Manual testing**: Always test manually with real credentials before committing

### Documentation
1. **Update docs immediately**: Don't leave docs for later
2. **Use examples**: Show real commands users will run
3. **Document gotchas**: Mention region availability, IAM permissions, etc.
4. **Keep CLAUDE.md current**: This is the source of truth for developers

### Security
1. **Never hardcode credentials**: Use environment variables or config files
2. **Mask sensitive data**: Hide API keys in logs
3. **File permissions**: Config files should be 0600
4. **Validate inputs**: Check all user inputs before sending to API

---

## Common Pitfalls

### 1. Forgetting to Update Model Registry
❌ **Wrong**: Add client but forget to update `models.go`
✅ **Right**: Add model metadata to `AvailableModels()` immediately

### 2. Hardcoding Defaults
❌ **Wrong**: `region := "us-east-1"` (hardcoded)
✅ **Right**: Check env vars, config file, then default

### 3. Ignoring Rate Limits
❌ **Wrong**: No circuit breaker, retry forever
✅ **Right**: Use circuit breaker, exponential backoff, max retries

### 4. Poor Error Messages
❌ **Wrong**: `return fmt.Errorf("error: %w", err)`
✅ **Right**: `return fmt.Errorf("authentication failed: check AWS_ACCESS_KEY_ID and IAM permissions: %w", err)`

### 5. Incomplete Testing
❌ **Wrong**: Only test happy path
✅ **Right**: Test errors, edge cases, retries, circuit breaker

### 6. Missing Documentation
❌ **Wrong**: Code works, ship it
✅ **Right**: Update all docs, add examples, create migration guide

---

## Support & Resources

### When You Get Stuck

1. **Check existing implementations**: Look at `gemini_rest.go` or `vertex_sdk.go` for patterns
2. **Review provider docs**: AWS/Google docs are comprehensive
3. **Test incrementally**: Build one method at a time, test as you go
4. **Use verbose mode**: `--verbose` flag shows detailed logs
5. **Ask for help**: Open an issue or discussion if blocked

### Useful References

- [AWS Bedrock Docs](https://docs.aws.amazon.com/bedrock/)
- [Google Gemini API Docs](https://ai.google.dev/docs)
- [Google Vertex AI Docs](https://cloud.google.com/vertex-ai/docs)
- [Go AWS SDK v2 Docs](https://aws.github.io/aws-sdk-go-v2/)
- [MCP Protocol Spec](https://modelcontextprotocol.io/)

---

## Conclusion

Adding a new model to gimage is systematic and well-structured. By following this guide, you can integrate any image generation API in a consistent, maintainable way. The multi-backend architecture makes it easy to support many providers while maintaining a clean, unified interface for users.

**Key Takeaways**:
- Research thoroughly before coding
- Follow existing patterns and conventions
- Test extensively (unit, integration, manual)
- Document everything immediately
- Security and error handling are critical

Good luck with your integration! 🚀
