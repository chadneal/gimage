# GImage Image Generation Backends - Comprehensive Audit

## Executive Summary

The gimage project implements a multi-backend image generation system supporting three cloud providers (Gemini, Vertex AI, AWS Bedrock). The architecture demonstrates good separation of concerns with a provider registry pattern, but reveals several code duplication issues and inconsistencies in error handling and authentication flows.

**Key Findings:**
- Well-designed provider registry system
- Significant code duplication across similar backends
- Inconsistent error handling patterns
- Complex authentication flows with multiple fallback options
- Opportunity for better abstraction of common REST/SDK patterns

---

## Architecture Overview

### Provider Registry Pattern

**Location:** `/home/user/gimage/internal/generate/providers.go`

The system uses a central `ProviderRegistry` that manages all available providers with metadata about:
- Authentication requirements (EnvVar struct with config keys)
- Pricing information
- Capabilities (styles, negative prompts, seed support)
- Client factory functions

```
┌─────────────────────────────────────────────────────────┐
│           ProviderRegistry                              │
├─────────────────────────────────────────────────────────┤
│ Providers:                                              │
│  • gemini/flash-2.5 → GeminiRESTClient                 │
│  • vertex/imagen-4 → VertexRESTClient or VertexSDK     │
│  • vertex/imagen-3 → VertexRESTClient or VertexSDK     │
│  • bedrock/nova-canvas → BedrockRESTClient or SDK      │
└─────────────────────────────────────────────────────────┘
```

**Strengths:**
- Decentralized client creation
- Provider aliasing (e.g., "gemini" → "gemini/flash-2.5")
- Clear metadata about requirements
- Authentication status tracking

---

## Backend Implementations

### 1. Gemini API (REST-only)

**File:** `gemini_rest.go` | **Interface:** `GeminiRESTClient`

```
Authentication: API Key (single method)
Protocol: REST (generateContent endpoint)
URL: https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
Models: gemini-2.5-flash-image (single model)
```

**Characteristics:**
- REST-only implementation
- Simple authentication (API key in query param)
- Base64 image encoding in response
- Comprehensive error classification for 401/403/429/5xx

**Request Structure:**
```json
{
  "contents": [{"parts": [{"text": "prompt"}]}],
  "generationConfig": {"responseModalities": ["IMAGE"]}
}
```

### 2. Vertex AI (Dual Implementation: REST + SDK)

**Files:** `vertex_rest.go`, `vertex_sdk.go`

```
Authentication Methods:
  - REST: API Key (x-goog-api-key header)
  - SDK: GOOGLE_APPLICATION_CREDENTIALS (service account)
  
Protocol: 
  - REST: Vertex AI Platform API
  - SDK: Vertex AI genai Go client

Models: imagen-4.0, imagen-3.0, imagen-3.0-fast variants (multiple models)
```

**Dual Implementation Complexity:**

| Aspect | REST | SDK |
|--------|------|-----|
| Setup | Simple (API key) | Complex (service account file) |
| Auth | x-goog-api-key header | GOOGLE_APPLICATION_CREDENTIALS |
| Init | HTTP client | genai.Client instance |
| Models | All available | Requires proper permissions |
| Aspect Ratio | Manual calculation from dimensions | Not exposed in interface |

**Request Structure (REST):**
```json
{
  "instances": [{"prompt": "..."}],
  "parameters": {
    "sampleCount": 1,
    "aspectRatio": "1:1",
    "negativePrompt": "...",
    "seed": 12345
  }
}
```

### 3. AWS Bedrock (Dual Implementation: REST + SDK)

**Files:** `bedrock_rest.go`, `bedrock_sdk.go`

```
Authentication Methods:
  - REST: Bearer token (AWS_BEDROCK_API_KEY)
  - SDK: AWS credentials (keys, profile, IAM role)

Protocol:
  - REST: Bedrock Runtime HTTP API
  - SDK: AWS SDK v2 (aws-sdk-go-v2)

Models: amazon.nova-canvas-v1:0 (single model)
```

**Dual Implementation Similarities:**

Both implementations share:
- Same `NovaCanvasRequest` struct definition
- Identical dimension validation (512-2048, multiple of 64)
- Identical seed validation (0-858993459)
- Same quality mapping logic
- Same base64 image decoding

---

## Critical Code Duplication Analysis

### Pattern 1: Shared Request Building Logic

**Duplication:** `buildRequest()` method exists identically in:
- `bedrock_rest.go` (lines 203-261)
- `bedrock_sdk.go` (lines 211-269)

```go
// Identical code appears in both files:
func (c *BedrockSDKClient) buildRequest(prompt string, options models.GenerateOptions) (*NovaCanvasRequest, error) {
    // Lines 203-261 duplicated exactly in bedrock_rest.go
}
```

**Impact:** Bug fixes need to be applied in two places. Test duplication.

### Pattern 2: Error Handling Duplication

**Duplication:** Error response parsing logic:

**gemini_rest.go (lines 285-312):**
```go
func (c *GeminiRESTClient) handleHTTPError(statusCode int, body []byte) error {
    // Parse error response
    var errorResp struct { Error struct { Code int; Message string; Status string } }
    // Status code switch: 401, 403, 429, 500/502/503
}
```

**vertex_rest.go (lines 330-360):**
```go
func (c *VertexRESTClient) handleHTTPError(statusCode int, body []byte) error {
    // IDENTICAL error parsing logic
    // IDENTICAL status code handling
    // Provider-specific messages differ slightly
}
```

**bedrock_rest.go (lines 264-312):**
```go
func (c *BedrockRESTClient) handleHTTPError(statusCode int, body []byte) error {
    // Similar structure but uses http.StatusConsts
    // Bedrock-specific messages
}
```

**Impact:** 
- 3+ copies of nearly identical error handling
- Inconsistent error message format
- HTTP status code handling duplicated
- Different logging strategies (fmt vs zerolog)

### Pattern 3: Circuit Breaker Wrapping

**Duplication:** Pattern appears in all REST clients:

```go
// In GeminiRESTClient.GenerateImage()
for attempt := 1; attempt <= maxRetries; attempt++ {
    result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
        return c.generateWithRetry(ctx, modelName, enhancedPrompt, options)
    })
    // Check circuit breaker error
    // Check retryable error
    // Exponential backoff logic
}

// In VertexRESTClient.GenerateImage()
// IDENTICAL structure

// In BedrockRESTClient.GenerateImage()
// IDENTICAL structure (circuit breaker logic inside lambda)
```

**Impact:** Retry logic is duplicated across 3 REST clients. Difficult to maintain consistent retry behavior.

### Pattern 4: Verbose Logging Setup

**Duplication:** Every client initializes verbose logging identically:

```go
// gemini_rest.go
verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true" || os.Getenv("VERBOSE") == "true"

// vertex_rest.go
// IDENTICAL

// bedrock_rest.go
verbose := viper.GetBool("verbose") ||
    os.Getenv("GIMAGE_VERBOSE") == "true" ||
    os.Getenv("VERBOSE") == "true"
```

### Pattern 5: Dimension Parsing

**Duplication:** `parseDimensions()` function defined in gemini_rest.go but used by all providers:

```go
// Used by: gemini_rest.go, vertex_rest.go, bedrock_rest.go, bedrock_sdk.go
// Yet only defined once, requires imports or re-implementations
```

Actually checking code... it appears `parseDimensions()` is defined in `gemini_rest.go` (lines 368-389) and re-used by other files without duplication (calls the function directly). This is actually a good pattern.

---

## Authentication Complexity

### Authentication Matrix

| Provider | Auth Method 1 | Auth Method 2 | Fallback | Complexity |
|----------|---------------|---------------|----------|-----------|
| **Gemini** | API Key | None | None | Low |
| **Vertex** | API Key (REST) | Service Account (SDK) | SDK if no API key | Medium |
| **Bedrock** | Bearer Token | AWS Keys | SDK with credential chain | High |

### Vertex AI Authentication Flow

**File:** `providers.go` (lines 180-195)

```go
CreateClient: func(creds map[string]string) (ImageGenerator, error) {
    project := creds["VERTEX_PROJECT"]
    location := creds["VERTEX_LOCATION"]
    apiKey := creds["VERTEX_API_KEY"]
    
    if project == "" || location == "" {
        return nil, fmt.Errorf("VERTEX_PROJECT and VERTEX_LOCATION are required")
    }
    
    if apiKey != "" {
        return NewVertexRESTClient(apiKey, project, location)
    }
    ctx := context.Background()
    return NewVertexSDKClient(ctx, project, location)
}
```

**Issues:**
1. Context created inside factory with `context.Background()` - not ideal
2. Complex fallback: REST → SDK
3. REST requires: VERTEX_API_KEY + VERTEX_PROJECT + VERTEX_LOCATION
4. SDK requires: GOOGLE_APPLICATION_CREDENTIALS + VERTEX_PROJECT + VERTEX_LOCATION
5. If both unset, SDK attempt will fail with generic error

### Bedrock Authentication Flow

**File:** `providers.go` (lines 419-434)

```go
CreateClient: func(creds map[string]string) (ImageGenerator, error) {
    region := creds["AWS_REGION"]
    if region == "" {
        return nil, fmt.Errorf("AWS_REGION is required")
    }
    
    // Try bearer token first
    if bearerToken := creds["AWS_BEDROCK_API_KEY"]; bearerToken != "" {
        return NewBedrockRESTClient(bearerToken, region)
    }
    
    // Fall back to SDK
    ctx := context.Background()
    return NewBedrockSDKClient(ctx, region)
}
```

**Issues:**
1. Bearer token approach unusual for AWS (not standard AWS auth)
2. Multiple credential sources: custom bearer token OR AWS SDK credential chain
3. No clear preference signal to user about which auth method will be used
4. Fallback happens silently if bearer token not set

---

## Error Handling Patterns

### Inconsistency: Error Classification

**Gemini (gemini_rest.go):**
```go
func isRetryableError(err error) bool {
    if err == nil { return false }
    errStr := err.Error()
    retryablePatterns := []string{
        "rate limit", "quota exceeded", "timeout", "deadline exceeded",
        "connection", "unavailable", "503", "429",
    }
    for _, pattern := range retryablePatterns {
        if contains(errStr, pattern) { return true }
    }
    return false
}
```

**Bedrock SDK (bedrock_sdk.go):**
```go
func (c *BedrockSDKClient) handleError(err error) error {
    switch {
    case contains(errMsg, "ValidationException"): // AWS error types
    case contains(errMsg, "AccessDeniedException"):
    case contains(errMsg, "ThrottlingException"):
    case contains(errMsg, "ModelNotReadyException"):
    // ... different pattern matching
    }
}
```

**Issues:**
- String-based pattern matching fragile
- AWS SDK returns specific error types that could be type-asserted
- Different error codes between providers not unified
- No central error classification system

### Inconsistency: Context Cancellation

**Gemini (lines 96-97):**
```go
select {
case <-ctx.Done():
    return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
case <-time.After(backoff):
}
```

**Bedrock REST:** No context cancellation check during retry backoff
**Vertex REST:** No context cancellation check during retry backoff

**Impact:** Gemini respects context timeouts properly, others may hang longer than needed

---

## Logging Inconsistency

### Different Logging Strategies

**Gemini & Vertex REST (fmt-based):**
```go
func (c *GeminiRESTClient) logVerbose(format string, args ...interface{}) {
    if c.verbose {
        fmt.Fprintf(os.Stderr, "[GEMINI-REST] "+format+"\n", args...)
    }
}
```

**Bedrock (zerolog-based):**
```go
logger := zerolog.New(os.Stderr).With().
    Timestamp().
    Str("component", "bedrock_rest").
    Logger()

if c.verbose {
    c.logger.Debug().Str("prompt", prompt).Msg("Generating image...")
}
```

**Issues:**
- Inconsistent logging backends (fmt vs zerolog)
- Different log levels: Bedrock uses structured logging, others use simple printf
- Bedrock has timestamp support, others don't
- Different verbosity control mechanisms

---

## API Response Parsing Differences

### Response Structure Inconsistency

**Gemini (REST generateContent format):**
```go
type geminiGenerateContentResponse struct {
    Candidates []geminiCandidate `json:"candidates"`
}
type geminiCandidate struct {
    Content genai.Content `json:"content"`
    // Image data in: candidate.Content.Parts[i].InlineData.Data (base64)
}
```

**Vertex (predict format):**
```go
type vertexPredictResponse struct {
    Predictions []vertexPrediction `json:"predictions"`
}
type vertexPrediction struct {
    BytesBase64Encoded string `json:"bytesBase64Encoded"`
    MimeType          string `json:"mimeType"`
}
```

**Bedrock (Nova Canvas format):**
```go
type NovaCanvasResponse struct {
    Images []string `json:"images"`
    Error  string   `json:"error,omitempty"`
}
// Image data is array of base64 strings
```

**Impact:** Response parsing logic completely different for each provider, difficult to abstract

---

## Capability Matrix

### Feature Support Variations

| Feature | Gemini | Vertex Imagen-4 | Bedrock Nova |
|---------|--------|-------------------|---------------|
| Styles | Yes | Yes | Via quality enum |
| Negative Prompt | Yes | Yes | Yes |
| Seed Support | Yes | Yes | Yes |
| Max Prompt Length | 480 | 2000 | 4096 |
| Aspect Ratio Control | Yes | Yes (manual) | No |
| Quality/Premium | No | No | Yes |
| Number of Images | Always 1 | Always 1 | Always 1 |
| Custom CFG Scale | No | No | Yes (fixed 7.0) |

**Issues:**
- Provider capabilities very different
- CLI needs to handle these differences
- Some features silently dropped for certain providers

---

## Positive Patterns

### 1. Interface Consistency

All backends implement common `ImageGenerator` interface:
```go
type ImageGenerator interface {
    GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error)
    Close() error
}
```

### 2. Prompt Validation

Consistent prompt validation across all providers:
```go
// Every backend calls ValidatePrompt() before processing
if err := ValidatePrompt(prompt); err != nil {
    return nil, err
}
```

### 3. Circuit Breaker Pattern

All backends use circuit breaker for resilience:
- Configurable failure thresholds
- State transitions logged
- Prevents cascading failures

### 4. Provider Metadata

Rich metadata about each provider:
- Required environment variables
- Pricing information
- Capability descriptions
- Clear authentication requirements

---

## Summary of Key Issues

### High Priority

1. **Code Duplication in Error Handling**
   - Lines 285-312 (Gemini) ≈ Lines 330-360 (Vertex)
   - HTTP error parsing duplicated across REST clients
   - **Solution:** Extract to shared `handleHTTPError()` function

2. **Bedrock buildRequest() Duplication**
   - Identical implementation in REST (lines 203-261) and SDK (lines 211-269)
   - **Solution:** Extract to shared helper function

3. **Retry Logic Duplication**
   - Circuit breaker + exponential backoff pattern identical across 3 clients
   - **Solution:** Extract to shared retry wrapper function

4. **Inconsistent Error Handling**
   - Some providers check context cancellation, others don't
   - Different string-matching patterns for error classification
   - **Solution:** Implement centralized error classifier

### Medium Priority

5. **Logging Inconsistency**
   - Gemini/Vertex: fmt-based logging
   - Bedrock: zerolog-based logging
   - **Solution:** Standardize on one logging backend

6. **Authentication Complexity**
   - Vertex and Bedrock have fallback authentication flows
   - Context created in factory functions
   - **Solution:** Clearer authentication preference system

7. **Response Parsing Differences**
   - Each provider has different response structures
   - No common abstraction for image extraction
   - **Solution:** Create provider-specific response parsers

### Lower Priority

8. **Aspect Ratio Calculation**
   - Vertex REST manually calculates from dimensions
   - Gemini and Bedrock use native aspect ratio parameter
   - **Solution:** Consistent handling across providers

9. **Configuration Duplication**
   - Verbose flag check duplicated across clients
   - Default model and size hardcoded in multiple places
   - **Solution:** Pass configuration through client initialization

---

## Recommended Refactoring

### 1. Extract Common REST Client Base

```go
type RESTClientBase struct {
    httpClient     *http.Client
    verbose        bool
    circuitBreaker *gobreaker.CircuitBreaker
    logVerbose(string, ...interface{})
    handleHTTPError(int, []byte) error
    withRetry(ctx, func) error  // Shared retry logic
}
```

### 2. Error Classification System

```go
type APIError struct {
    Provider string
    Code     int
    Message  string
    Type     ErrorType // RETRYABLE, PERMANENT, RATE_LIMITED, etc.
}

func classifyError(provider string, err error, statusCode int) APIError
```

### 3. Provider Response Adapter Pattern

```go
type ResponseAdapter interface {
    ExtractImage(rawResponse []byte) ([]byte, string, error) // image data, mime type, error
    ExtractMetadata(rawResponse []byte) map[string]string
}
```

### 4. Authentication Factory

```go
type AuthProvider interface {
    Authenticate(ctx context.Context) error
    Type() string // "api_key", "service_account", "aws_credentials"
}

type ClientFactory struct {
    auth AuthProvider
    config ClientConfig
}
```

---

## Testing Observations

### Test Coverage
- Unit tests exist for utilities (prompt, download, circuitbreaker)
- Basic client initialization tests (gemini_test.go, bedrock_rest_test.go)
- No integration tests included (expected - would require credentials)

### Test Gaps
- No tests for error handling branches
- No tests for retry logic
- No tests for circuit breaker behavior
- No tests for response parsing
- No tests for dimension/aspect ratio calculations

---

## Conclusion

The gimage backend system demonstrates solid architectural foundations with the provider registry pattern and consistent interface design. However, significant code duplication in error handling, retry logic, and request building creates maintenance burden and test coverage gaps.

Priority improvements should focus on:
1. Extracting shared error handling
2. Consolidating retry logic
3. Unifying logging strategy
4. Simplifying authentication flows

These changes would reduce code complexity by ~20-30%, improve test coverage, and make the system more maintainable as new providers are added.

