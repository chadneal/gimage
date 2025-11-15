package gimage_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	gimage "github.com/apresai/gimage/sdk/go"
)

// Example demonstrates how to use the Gimage Go SDK
func Example() {
	// Create a client with your API Gateway endpoint
	baseURL := "https://cf3xrk9w63.execute-api.us-east-1.amazonaws.com/production"

	client, err := gimage.NewClient(baseURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create context
	ctx := context.Background()

	// Generate an image
	resp, err := client.GenerateImage(ctx, gimage.GenerateImageJSONRequestBody{
		Prompt: "sunset over mountains",
		Model:  stringPtr("gemini-2.5-flash-image"),
		Size:   stringPtr("1024x1024"),
	})
	if err != nil {
		log.Fatalf("Failed to generate image: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
}

// ExampleWithAPIKey demonstrates authentication with API Gateway API key
func ExampleWithAPIKey() {
	baseURL := "https://cf3xrk9w63.execute-api.us-east-1.amazonaws.com/production"
	apiKey := "your-api-key-here"

	// Create client with API key authentication
	client, err := gimage.NewClient(baseURL, gimage.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Generate image with custom options
	resp, err := client.GenerateImage(ctx, gimage.GenerateImageJSONRequestBody{
		Prompt:         "futuristic city with flying cars",
		Model:          stringPtr("gemini-2.5-flash-image"),
		Size:           stringPtr("1024x1024"),
		Style:          (*gimage.ImageStyle)(stringPtr("photorealistic")),
		NegativePrompt: stringPtr("people, text"),
		Seed:           intPtr(42),
		ResponseFormat: (*gimage.ResponseFormat)(stringPtr("base64")),
	})
	if err != nil {
		log.Fatalf("Failed to generate image: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
}

// ExampleHealthCheck demonstrates checking API health
func ExampleHealthCheck() {
	baseURL := "https://cf3xrk9w63.execute-api.us-east-1.amazonaws.com/production"
	apiKey := "your-api-key-here"

	client, err := gimage.NewClient(baseURL, gimage.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Check health
	resp, err := client.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Health Status: %d\n", resp.StatusCode)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
