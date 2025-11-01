package generate

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
	"github.com/apresai/gimage/pkg/models"
	"github.com/sony/gobreaker"
	"github.com/spf13/viper"
)

// VertexSDKClient uses Vertex AI SDK for image generation
// This requires GOOGLE_APPLICATION_CREDENTIALS environment variable
// pointing to a service account JSON key file
type VertexSDKClient struct {
	client         *genai.Client
	project        string
	location       string
	verbose        bool
	circuitBreaker *gobreaker.CircuitBreaker
}

// NewVertexSDKClient creates a new Vertex AI SDK client
// It uses the GOOGLE_APPLICATION_CREDENTIALS environment variable for authentication
func NewVertexSDKClient(ctx context.Context, project, location string) (*VertexSDKClient, error) {
	// Check for service account credentials
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable not set\n" +
			"Hint: Set up Vertex AI credentials using: export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json")
	}

	// Verify credentials file exists
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file not found: %s", credsPath)
	}

	// Get project from parameter or environment
	if project == "" {
		project = os.Getenv("VERTEX_PROJECT")
		if project == "" {
			return nil, fmt.Errorf("project ID required: set VERTEX_PROJECT environment variable or use --project flag")
		}
	}

	// Default location
	if location == "" {
		location = "us-central1"
	}

	// Check if verbose mode is enabled
	verbose := viper.GetBool("verbose") || os.Getenv("GIMAGE_VERBOSE") == "true" || os.Getenv("VERBOSE") == "true"

	if verbose {
		fmt.Fprintf(os.Stderr, "[VERTEX-SDK] Using credentials from: %s\n", credsPath)
		fmt.Fprintf(os.Stderr, "[VERTEX-SDK] Project: %s, Location: %s\n", project, location)
	}

	// Create Vertex AI client
	client, err := genai.NewClient(ctx, project, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w\nHint: Check that the service account has 'Vertex AI User' role", err)
	}

	return &VertexSDKClient{
		client:         client,
		project:        project,
		location:       location,
		verbose:        verbose,
		circuitBreaker: newCircuitBreaker("VertexSDK"),
	}, nil
}

// logVerbose logs debug information if verbose mode is enabled
func (c *VertexSDKClient) logVerbose(format string, args ...interface{}) {
	if c.verbose {
		fmt.Fprintf(os.Stderr, "[VERTEX-SDK] "+format+"\n", args...)
	}
}

// GenerateImage generates an image using Vertex AI SDK
func (c *VertexSDKClient) GenerateImage(ctx context.Context, prompt string, options models.GenerateOptions) (*models.GeneratedImage, error) {
	// Validate prompt
	if err := ValidatePrompt(prompt); err != nil {
		return nil, err
	}

	// Enhance prompt for better results
	enhancedPrompt := EnhancePrompt(prompt)

	// Use custom model if provided, otherwise default
	modelName := options.Model
	if modelName == "" {
		modelName = "imagen-4.0-generate-001"
	}

	// Format model name for Vertex AI (needs publishers/google/models/ prefix)
	fullModelName := fmt.Sprintf("publishers/google/models/%s", modelName)

	c.logVerbose("Generating image with model: %s", fullModelName)
	c.logVerbose("Prompt: %s", enhancedPrompt)

	// Get the generative model
	model := c.client.GenerativeModel(fullModelName)

	// Configure temperature for better image generation
	model.SetTemperature(0.9)

	c.logVerbose("Sending request to Vertex AI...")

	// Generate content through circuit breaker
	result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return model.GenerateContent(ctx, genai.Text(enhancedPrompt))
	})

	if err != nil {
		// Check if circuit breaker is open
		if isCircuitBreakerError(err) {
			c.logVerbose("Circuit breaker is open, failing fast")
			return nil, fmt.Errorf("API circuit breaker is open (too many failures): %w", err)
		}
		c.logVerbose("Generation failed: %v", err)
		return nil, fmt.Errorf("failed to generate image: %w\nHint: Check that billing is enabled and you have Vertex AI User role", err)
	}

	resp := result.(*genai.GenerateContentResponse)

	// Check if we got candidates
	if len(resp.Candidates) == 0 {
		c.logVerbose("No candidates in response")
		return nil, fmt.Errorf("no image generated from prompt")
	}

	c.logVerbose("Got %d candidates", len(resp.Candidates))

	// Extract image data from response
	var imageData []byte
	var format string = "png"

	for i, cand := range resp.Candidates {
		c.logVerbose("Candidate %d has %d parts", i, len(cand.Content.Parts))

		for j, part := range cand.Content.Parts {
			c.logVerbose("Part %d type: %T", j, part)

			// Check if this part contains image data (Blob type)
			if blob, ok := part.(genai.Blob); ok {
				c.logVerbose("Found Blob: mime=%s, size=%d bytes", blob.MIMEType, len(blob.Data))
				imageData = blob.Data

				// Determine format from MIME type
				switch blob.MIMEType {
				case "image/png":
					format = "png"
				case "image/jpeg", "image/jpg":
					format = "jpg"
				case "image/webp":
					format = "webp"
				}
				break
			}
		}

		if len(imageData) > 0 {
			break
		}
	}

	if len(imageData) == 0 {
		c.logVerbose("No image data found in response")
		return nil, fmt.Errorf("no image data found in response")
	}

	c.logVerbose("Successfully generated image: %d bytes, format=%s", len(imageData), format)

	// Parse dimensions from options
	width, height := parseDimensions(options.Size)

	return &models.GeneratedImage{
		Data:   imageData,
		Format: format,
		Width:  width,
		Height: height,
		Metadata: map[string]string{
			"model":    modelName,
			"prompt":   prompt,
			"style":    options.Style,
			"api":      "vertex-sdk",
			"project":  c.project,
			"location": c.location,
		},
	}, nil
}

// Close closes the Vertex AI client
func (c *VertexSDKClient) Close() error {
	return c.client.Close()
}
