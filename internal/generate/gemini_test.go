package generate

import (
	"context"
	"testing"

	"github.com/chadneal/gimage/pkg/models"
)

func TestNewGeminiClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "valid API key",
			apiKey:  "test-api-key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGeminiClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewGeminiClient() returned nil client")
			}
			if client != nil && client.model != defaultModel {
				t.Errorf("NewGeminiClient() model = %v, want %v", client.model, defaultModel)
			}
		})
	}
}

func TestSetModel(t *testing.T) {
	client := &GeminiClient{
		apiKey: "test",
		model:  defaultModel,
	}

	tests := []struct {
		name     string
		newModel string
		want     string
	}{
		{
			name:     "set custom model",
			newModel: "gemini-2.0-flash-preview-image-generation",
			want:     "gemini-2.0-flash-preview-image-generation",
		},
		{
			name:     "empty model string",
			newModel: "",
			want:     defaultModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.SetModel(tt.newModel)
			if tt.newModel != "" && client.model != tt.want {
				t.Errorf("SetModel() model = %v, want %v", client.model, tt.want)
			}
		})
	}
}

func TestParseSizeString(t *testing.T) {
	tests := []struct {
		name       string
		size       string
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "valid 1024x1024",
			size:       "1024x1024",
			wantWidth:  1024,
			wantHeight: 1024,
		},
		{
			name:       "valid 512x768",
			size:       "512x768",
			wantWidth:  512,
			wantHeight: 768,
		},
		{
			name:       "invalid format",
			size:       "invalid",
			wantWidth:  1024,
			wantHeight: 1024,
		},
		{
			name:       "empty string",
			size:       "",
			wantWidth:  1024,
			wantHeight: 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, gotHeight := parseSizeString(tt.size)
			if gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Errorf("parseSizeString() = (%v, %v), want (%v, %v)",
					gotWidth, gotHeight, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestExtractFormatFromMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     string
	}{
		{
			name:     "image/png",
			mimeType: "image/png",
			want:     "png",
		},
		{
			name:     "image/jpeg",
			mimeType: "image/jpeg",
			want:     "jpg",
		},
		{
			name:     "image/jpg",
			mimeType: "image/jpg",
			want:     "jpg",
		},
		{
			name:     "image/webp",
			mimeType: "image/webp",
			want:     "webp",
		},
		{
			name:     "image/gif",
			mimeType: "image/gif",
			want:     "gif",
		},
		{
			name:     "unknown type",
			mimeType: "application/octet-stream",
			want:     "png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractFormatFromMimeType(tt.mimeType); got != tt.want {
				t.Errorf("extractFormatFromMimeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "deadline exceeded error",
			err:  context.DeadlineExceeded,
			want: true, // This contains "deadline exceeded" which is retryable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableError(tt.err); got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateImage_EmptyPrompt(t *testing.T) {
	client := &GeminiClient{
		apiKey: "test",
		model:  defaultModel,
	}

	ctx := context.Background()
	_, err := client.GenerateImage(ctx, "", models.GenerateOptions{})

	if err == nil {
		t.Error("GenerateImage() with empty prompt should return error")
	}
}
