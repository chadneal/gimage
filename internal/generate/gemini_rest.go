package generate

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apresai/gimage/pkg/models"
	"github.com/sony/gobreaker"
	"github.com/spf13/viper"
)

// Gemini REST API endpoint
const geminiAPIEndpoint = "https://generativelanguage.googleapis.com/v1beta/models"

// GeminiRESTClient uses Gemini REST API for image generation
type GeminiRESTClient struct {
	apiKey         string
	model          string
	httpClient     *http.Client
	verbose        bool
	circuitBreaker *gobreaker.CircuitBreaker
}

// NewGeminiRESTClient creates a new Gemini REST API client
func NewGeminiRESTClient(apiKey string) (*GeminiRESTClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Check if verbose mode is enabled via Viper flag or environment variable
	verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true" || os.Getenv("VERBOSE") == "true"

	return &GeminiRESTClient{
		apiKey:  apiKey,
		model:   DefaultModel,
		verbose: verbose,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		circuitBreaker: newCircuitBreaker("GeminiAPI"),
	}, nil
}

// logVerbose logs debug information if verbose mode is enabled
func (c *GeminiRESTClient) logVerbose(format string, args ...interface{}) {
	if c.verbose {
		fmt.Fprintf(os.Stderr, "[GEMINI-REST] "+format+"\n", args...)
	}
}

// GenerateImage generates an image using Gemini REST API
func (c *GeminiRESTClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
	// Validate prompt
	if err := ValidatePrompt(prompt); err != nil {
		return nil, err
	}

	// Enhance prompt for better results
	enhancedPrompt := EnhancePrompt(prompt)

	// Use custom model if provided
	modelName := c.model
	if options.Model != "" {
		modelName = options.Model
	}

	// Generate image with circuit breaker and retry logic
	var lastErr error
	backoff := retryBackoffInitial

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Execute through circuit breaker
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
				// Exponential backoff with cap
				backoff *= 2
				if backoff > retryBackoffMax {
					backoff = retryBackoffMax
				}
			}
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// generateWithRetry performs a single generation attempt using REST API
func (c *GeminiRESTClient) generateWithRetry(ctx context.Context, modelName, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
	// Build the prompt with options
	fullPrompt := buildPromptWithOptions(prompt, options)

	// Parse dimensions from options
	width, height := parseDimensions(options.Size)

	c.logVerbose("Building request for model: %s", modelName)
	c.logVerbose("Full prompt: %s", fullPrompt)
	c.logVerbose("Requested dimensions: %dx%d", width, height)

	// Build request payload using Gemini's generateContent API format
	// According to Gemini API documentation, the request should have:
	// {
	//   "contents": [{
	//     "parts": [{"text": "prompt text"}]
	//   }],
	//   "generationConfig": {
	//     "responseModalities": ["IMAGE"]
	//   }
	// }
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{
						Text: fullPrompt,
					},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			ResponseModalities: []string{"IMAGE"},
		},
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logVerbose("Request body: %s", string(requestBody))

	// Build API URL - use generateContent endpoint
	apiURL := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIEndpoint, modelName, c.apiKey)

	c.logVerbose("API URL: %s", strings.Replace(apiURL, c.apiKey, "***KEY***", -1))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	c.logVerbose("Sending request to Gemini API...")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logVerbose("Request failed: %v", err)
		return nil, enhanceError(err)
	}
	defer resp.Body.Close()

	c.logVerbose("Response status: %d %s", resp.StatusCode, resp.Status)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response (truncated)
	if len(body) > 500 {
		c.logVerbose("Response body (first 500 chars): %s...", string(body[:500]))
	} else {
		c.logVerbose("Response body: %s", string(body))
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	// Parse response using Gemini's generateContent response format
	var response geminiGenerateContentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logVerbose("Failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}

	// Validate response structure
	if len(response.Candidates) == 0 {
		c.logVerbose("No candidates in response")
		return nil, fmt.Errorf("no image generated from prompt")
	}

	candidate := response.Candidates[0]
	if candidate.Content.Parts == nil || len(candidate.Content.Parts) == 0 {
		c.logVerbose("No parts in candidate content")
		return nil, fmt.Errorf("no content parts in response")
	}

	// Find the image part (should have inline_data)
	var imageData []byte
	var mimeType string
	found := false

	for i, part := range candidate.Content.Parts {
		c.logVerbose("Part %d: has InlineData=%v", i, part.InlineData != nil)
		if part.InlineData != nil && part.InlineData.Data != "" {
			// Decode base64 image data
			imageData, err = base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				c.logVerbose("Failed to decode base64 from part %d: %v", i, err)
				return nil, fmt.Errorf("failed to decode base64 image data: %w", err)
			}
			mimeType = part.InlineData.MimeType
			found = true
			c.logVerbose("Found image data in part %d: %d bytes, mime=%s", i, len(imageData), mimeType)
			break
		}
	}

	if !found {
		c.logVerbose("No inline_data found in any parts")
		return nil, fmt.Errorf("no image data found in response")
	}

	// Determine format from MIME type
	format := "png"
	if mimeType != "" {
		switch mimeType {
		case "image/jpeg":
			format = "jpg"
		case "image/png":
			format = "png"
		case "image/webp":
			format = "webp"
		}
	}

	c.logVerbose("Successfully generated image: %d bytes, format=%s", len(imageData), format)

	return &models.GeneratedImage{
		Data:   imageData,
		Format: format,
		Width:  width,
		Height: height,
		Metadata: map[string]string{
			"model":  modelName,
			"prompt": prompt,
			"style":  options.Style,
			"api":    "gemini-rest",
		},
	}, nil
}

// handleHTTPError handles HTTP error responses from the API
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

// Close closes the client connection
func (c *GeminiRESTClient) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// Request/Response structs for Gemini REST API (generateContent format)
type geminiGenerateContentRequest struct {
	Contents         []geminiContent          `json:"contents"`
	GenerationConfig *geminiGenerationConfig  `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text       string           `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

type geminiGenerationConfig struct {
	ResponseModalities []string `json:"responseModalities,omitempty"`
	Temperature        *float64 `json:"temperature,omitempty"`
	TopP               *float64 `json:"topP,omitempty"`
	TopK               *int     `json:"topK,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content       geminiContent `json:"content"`
	FinishReason  string        `json:"finishReason,omitempty"`
	SafetyRatings []interface{} `json:"safetyRatings,omitempty"`
}

// buildPromptWithOptions enhances the prompt with style options
func buildPromptWithOptions(prompt string, options models.GenerateOptions) string {
	enhanced := prompt
	if options.Style != "" {
		enhanced = fmt.Sprintf("%s, %s style", enhanced, options.Style)
	}
	return enhanced
}

// parseDimensions parses "WIDTHxHEIGHT" string
func parseDimensions(size string) (int, int) {
	if size == "" {
		return 1024, 1024
	}

	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 1024, 1024
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil || width <= 0 {
		return 1024, 1024
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil || height <= 0 {
		return 1024, 1024
	}

	return width, height
}

// enhanceError provides more context for API errors
func enhanceError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for authentication errors
	if contains(errStr, "401") || contains(errStr, "unauthorized") {
		return fmt.Errorf("authentication failed: invalid API key. Set GEMINI_API_KEY or use --api-key flag: %w", err)
	}

	// Check for permission errors
	if contains(errStr, "403") || contains(errStr, "forbidden") {
		return fmt.Errorf("permission denied: check your API key and quota: %w", err)
	}

	// Check for rate limiting
	if contains(errStr, "429") || contains(errStr, "rate limit") {
		return fmt.Errorf("rate limit exceeded: too many requests. Try again later: %w", err)
	}

	// Check for server errors
	if contains(errStr, "500") || contains(errStr, "502") || contains(errStr, "503") {
		return fmt.Errorf("server error: Gemini API is experiencing issues. Try again later: %w", err)
	}

	return err
}
