# GImage Backend Code Duplication - Detailed Reference

## Duplication #1: Bedrock buildRequest() Method

### Location
- **REST:** `/home/user/gimage/internal/generate/bedrock_rest.go` (lines 202-261)
- **SDK:** `/home/user/gimage/internal/generate/bedrock_sdk.go` (lines 210-269)

### Code Comparison

Both files contain this exact method:

```go
func (c *Bedrock[REST|SDK]Client) buildRequest(prompt string, options models.GenerateOptions) (*NovaCanvasRequest, error) {
    // Validate prompt
    if prompt == "" {
        return nil, fmt.Errorf("prompt cannot be empty")
    }

    // Parse dimensions
    width, height := parseDimensions(options.Size)

    // Validate dimensions (Nova Canvas supports 512-2048, multiples of 64)
    if width < 512 || width > 2048 || width%64 != 0 {
        return nil, fmt.Errorf("invalid width: %d (must be 512-2048, multiple of 64)", width)
    }
    if height < 512 || height > 2048 || height%64 != 0 {
        return nil, fmt.Errorf("invalid height: %d (must be 512-2048, multiple of 64)", height)
    }

    // Validate seed if provided (Nova Canvas supports 0-858993459)
    if options.Seed < 0 || options.Seed > 858993459 {
        return nil, fmt.Errorf("invalid seed: %d (must be 0-858993459)", options.Seed)
    }

    // Determine quality from style
    quality := "standard"
    if options.Style != "" {
        lowerStyle := strings.ToLower(options.Style)
        if lowerStyle == "premium" || lowerStyle == "high" || lowerStyle == "ultra" || lowerStyle == "photorealistic" {
            quality = "premium"
        }
    }

    // Build request
    request := &NovaCanvasRequest{
        TaskType: "TEXT_IMAGE",
        TextToImageParams: NovaCanvasTextToImageParams{
            Text: prompt,
        },
        ImageGenerationConfig: NovaCanvasImageConfig{
            NumberOfImages: 1,
            Quality:        quality,
            Height:         height,
            Width:          width,
            CfgScale:       7.0,
        },
    }

    // Add negative prompt if provided
    if options.NegativePrompt != "" {
        request.TextToImageParams.NegativeText = options.NegativePrompt
    }

    // Add seed if provided
    if options.Seed != 0 {
        request.ImageGenerationConfig.Seed = int(options.Seed)
    }

    return request, nil
}
```

### Impact
- Bug fixes needed in two places
- Inconsistent changes between files
- Duplicated test coverage requirement
- Shared data structures (`NovaCanvasRequest`) but separate implementations

### Recommended Solution
Create shared function in `bedrock_sdk.go`:
```go
// buildNovaCanvasRequest shared by both REST and SDK clients
func buildNovaCanvasRequest(prompt string, options models.GenerateOptions) (*NovaCanvasRequest, error) {
    // ... shared implementation
}
```

---

## Duplication #2: Error Handling Pattern

### Location
- **Gemini:** `/home/user/gimage/internal/generate/gemini_rest.go` (lines 285-312)
- **Vertex:** `/home/user/gimage/internal/generate/vertex_rest.go` (lines 330-360)
- **Bedrock:** `/home/user/gimage/internal/generate/bedrock_rest.go` (lines 264-312)

### Gemini Implementation
```go
func (c *GeminiRESTClient) handleHTTPError(statusCode int, body []byte) error {
    // Try to parse error response
    var errorResp struct {
        Error struct {
            Code    int    `json:"code"`
            Message string `json:"message"`
            Status  string `json:"status"`
        } `json:"error"`
    }

    if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
        return fmt.Errorf("API error %d: %s", statusCode, errorResp.Error.Message)
    }

    // Generic error messages based on status code
    switch statusCode {
    case 401:
        return fmt.Errorf("authentication failed (401): invalid API key. Please check your GEMINI_API_KEY")
    case 403:
        return fmt.Errorf("permission denied (403): API key may not have access to image generation")
    case 429:
        return fmt.Errorf("rate limit exceeded (429): too many requests, please try again later")
    case 500, 502, 503:
        return fmt.Errorf("server error (%d): the API is temporarily unavailable, please retry", statusCode)
    default:
        return fmt.Errorf("HTTP error %d: %s", statusCode, string(body))
    }
}
```

### Vertex Implementation (Nearly Identical)
```go
func (c *VertexRESTClient) handleHTTPError(statusCode int, body []byte) error {
    // Try to parse error response
    var errorResp struct {
        Error struct {
            Code    int    `json:"code"`
            Message string `json:"message"`
            Status  string `json:"status"`
        } `json:"error"`
    }

    if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
        return fmt.Errorf("API error %d: %s", statusCode, errorResp.Error.Message)
    }

    // Generic error messages based on status code
    switch statusCode {
    case 401:
        return fmt.Errorf("authentication failed (401): invalid API key. Please check your VERTEX_API_KEY")
    case 403:
        return fmt.Errorf("permission denied (403): API key may not have access to Vertex AI or project")
    case 404:
        return fmt.Errorf("not found (404): check project ID (%s) and model name", c.projectID)
    case 429:
        return fmt.Errorf("rate limit exceeded (429): too many requests, please try again later")
    case 500, 502, 503:
        return fmt.Errorf("server error (%d): the API is temporarily unavailable, please retry", statusCode)
    default:
        return fmt.Errorf("HTTP error %d: %s", statusCode, string(body))
    }
}
```

### Bedrock Implementation (Different Style)
```go
func (c *BedrockRESTClient) handleHTTPError(statusCode int, body []byte) error {
    // Try to parse error message from response
    var errorMsg string
    var errorResponse map[string]interface{}
    if err := json.Unmarshal(body, &errorResponse); err == nil {
        if msg, ok := errorResponse["message"].(string); ok {
            errorMsg = msg
        }
    }

    if errorMsg == "" {
        errorMsg = string(body)
    }

    // Log raw error in verbose mode
    if c.verbose {
        c.logger.Error().
            Int("status_code", statusCode).
            Str("error_message", errorMsg).
            Msg("AWS Bedrock REST API error")
    }

    // Provide user-friendly messages based on status code
    switch statusCode {
    case http.StatusBadRequest:
        return fmt.Errorf("invalid request (400): %s\n\nTip: Check image dimensions (512-2048, multiple of 64) and quality (standard/premium)", errorMsg)
    // ... more cases
    }
}
```

### Differences
| Aspect | Gemini | Vertex | Bedrock |
|--------|--------|--------|---------|
| Error struct parsing | Custom struct | Custom struct | Generic map |
| Status codes | 401,403,429,5xx | 401,403,404,429,5xx | 400,401,403,404,429,500,503 |
| Logging | fmt (if verbose) | fmt (if verbose) | zerolog |
| Message format | "API error %d: %s" | "API error %d: %s" | "Custom message (code): %s\n\nTip: ..." |

---

## Duplication #3: Retry Logic Pattern

### Location
- **Gemini REST:** `/home/user/gimage/internal/generate/gemini_rest.go` (lines 60-119)
- **Vertex REST:** `/home/user/gimage/internal/generate/vertex_rest.go` (lines 75-134)

### Gemini Pattern
```go
func (c *GeminiRESTClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    // ... validation ...
    
    var lastErr error
    backoff := retryBackoffInitial

    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
            return c.generateWithRetry(ctx, modelName, enhancedPrompt, options)
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
```

### Vertex Pattern (IDENTICAL STRUCTURE)
```go
func (c *VertexRESTClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    // ... validation ...
    
    var lastErr error
    backoff := retryBackoffInitial

    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
            return c.generateWithRetry(ctx, modelName, enhancedPrompt, options)
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
```

### Bedrock REST (Similar but Slightly Different)
```go
func (c *BedrockRESTClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
    startTime := time.Now()

    // ... build request ...

    // Execute request via circuit breaker
    var response *models.GeneratedImage
    _, err = c.circuitBreaker.Execute(func() (interface{}, error) {
        // Make HTTP request
        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, fmt.Errorf("HTTP request failed: %w", err)
        }
        defer resp.Body.Close()

        // Read response body
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("failed to read response: %w", err)
        }

        // Check for HTTP errors
        if resp.StatusCode != http.StatusOK {
            return nil, c.handleHTTPError(resp.StatusCode, body)
        }

        // Parse response
        // ... build response ...

        return response, nil
    })

    if err != nil {
        return nil, err
    }

    return response, nil
}
```

**Note:** Bedrock REST doesn't have retry logic outside circuit breaker! This is an inconsistency.

---

## Duplication #4: Verbose Flag Initialization

### Location
All files in `/home/user/gimage/internal/generate/`:

**Gemini REST** (lines 39-40):
```go
verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true" || os.Getenv("VERBOSE") == "true"
```

**Vertex REST** (lines 52-53):
```go
verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true" || os.Getenv("VERBOSE") == "true"
```

**Bedrock REST** (lines 49-52):
```go
verbose := viper.GetBool("verbose") ||
    os.Getenv("GIMAGE_VERBOSE") == "true" ||
    os.Getenv("VERBOSE") == "true"
```

**Bedrock SDK** (lines 84-86):
```go
verbose := viper.GetBool("verbose") ||
    os.Getenv("GIMAGE_VERBOSE") == "true" ||
    os.Getenv("VERBOSE") == "true"
```

**Vertex SDK** (lines 54):
```go
verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true" || os.Getenv("VERBOSE") == "true"
```

### Recommended Solution
Create utility function in `generate.go`:
```go
func isVerboseMode() bool {
    return viper.GetBool("verbose") || 
           os.Getenv("GIMAGE_VERBOSE") == "true" || 
           os.Getenv("VERBOSE") == "true"
}
```

---

## Duplication #5: Context Creation in Factories

### Location
- **Vertex:** `/home/user/gimage/internal/generate/providers.go` (lines 192-193, 250-251, 308-309, 366-367)
- **Bedrock:** `/home/user/gimage/internal/generate/providers.go` (lines 431-432)

### Code
```go
// Vertex - Called 4 times in provider factory
ctx := context.Background()
return NewVertexSDKClient(ctx, project, location)

// Bedrock - Called once
ctx := context.Background()
return NewBedrockSDKClient(ctx, region)
```

### Issue
- Context passed to factory is not ideal (should be from CLI)
- `context.Background()` doesn't support timeout
- Called in multiple places, duplicated

### Recommended Solution
Pass context to factory method:
```go
func (r *ProviderRegistry) CreateClientWithContext(ctx context.Context, providerID string) (ImageGenerator, error) {
    // ... implementation using provided context
}
```

---

## Duplication #6: Model Default Hardcoding

### Locations
1. **providers.go** (line 17): `const DefaultModel = "gemini-2.5-flash-image"`
2. **gemini_rest.go** (line 44): `model: DefaultModel,`
3. **vertex_rest.go** (line 59): `model: "imagen-4.0-generate-001",` (hardcoded!)
4. **vertex_sdk.go** (line 96): `modelName = "imagen-4.0-generate-001"` (hardcoded!)
5. **bedrock_rest.go** (line 104): `modelID := "amazon.nova-canvas-v1:0"` (hardcoded!)
6. **bedrock_sdk.go** (line 134): `modelID := "amazon.nova-canvas-v1:0"` (hardcoded!)

### Files
- `/home/user/gimage/internal/generate/providers.go`
- `/home/user/gimage/internal/generate/gemini_rest.go`
- `/home/user/gimage/internal/generate/vertex_rest.go`
- `/home/user/gimage/internal/generate/vertex_sdk.go`
- `/home/user/gimage/internal/generate/bedrock_rest.go`
- `/home/user/gimage/internal/generate/bedrock_sdk.go`

### Solution
Create constants in each module:
```go
const (
    defaultGeminiModel  = "gemini-2.5-flash-image"
    defaultVertexModel  = "imagen-4.0-generate-001"
    defaultBedrockModel = "amazon.nova-canvas-v1:0"
)
```

---

## Summary Table

| Duplication | Files | Lines | Type | Severity |
|------------|-------|-------|------|----------|
| buildRequest() | 2 files | ~60 lines | Method duplication | High |
| handleHTTPError() | 3 files | ~80 lines | Function duplication | High |
| Retry logic | 2-3 files | ~60 lines | Pattern duplication | High |
| Verbose initialization | 5 files | ~5 lines each | Code snippet | Medium |
| Model defaults | 6 locations | ~1 line each | Constant duplication | Medium |
| Context creation | 5 locations | ~1 line each | Code snippet | Low |

