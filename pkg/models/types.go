package models

// GenerateOptions contains configuration for AI image generation
type GenerateOptions struct {
	Model          string
	Size           string
	AspectRatio    string
	Style          string
	NegativePrompt string
	Seed           int64
}

// GeneratedImage represents the result of an AI image generation request
type GeneratedImage struct {
	Data     []byte
	Format   string
	Width    int
	Height   int
	Metadata map[string]string
}
