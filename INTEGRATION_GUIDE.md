# Gimage Lambda API - Integration Guide

Complete guide for integrating the Gimage Lambda API into your applications.

## Table of Contents

- [Quick Start](#quick-start)
- [API Overview](#api-overview)
- [Authentication](#authentication)
- [Client SDKs](#client-sdks)
  - [TypeScript/JavaScript](#typescriptjavascript)
  - [Python](#python)
  - [Go](#go)
  - [cURL](#curl)
- [Common Integration Patterns](#common-integration-patterns)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Rate Limits & Quotas](#rate-limits--quotas)
- [Examples by Use Case](#examples-by-use-case)

---

## Quick Start

### 1. Get Your API Endpoint

After deploying with `make deploy-lambda`, you'll receive an API Gateway URL:

```
https://abcdef1234.execute-api.us-east-1.amazonaws.com/prod
```

### 2. Test the API

```bash
curl https://YOUR_API_URL/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "0.1.1",
  "apis": {
    "gemini": "available",
    "vertex": "not_configured"
  }
}
```

### 3. Generate Your First Image

```bash
curl -X POST https://YOUR_API_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "a sunset over mountains",
    "size": "1024x1024",
    "response_format": "s3_url"
  }'
```

---

## API Overview

### Base URL

```
https://YOUR_API_ID.execute-api.REGION.amazonaws.com/prod
```

### Available Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `POST` | `/generate` | Generate AI images from text |
| `POST` | `/resize` | Resize to specific dimensions |
| `POST` | `/scale` | Scale by factor |
| `POST` | `/crop` | Crop to region |
| `POST` | `/compress` | Compress with quality |
| `POST` | `/convert` | Convert format |
| `POST` | `/batch` | Process multiple images |
| `GET` | `/health` | Health check |

### Response Formats

The API supports two response modes based on image size:

**Base64 (for small images < 512KB)**
```json
{
  "image": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB...",
  "width": 800,
  "height": 600,
  "format": "png",
  "size_bytes": 102400
}
```

**S3 URL (for large images)**
```json
{
  "s3_url": "https://s3.amazonaws.com/bucket/key?presigned...",
  "s3_key": "images/1234567890-abc.png",
  "width": 2048,
  "height": 2048,
  "format": "png",
  "size_bytes": 2097152
}
```

---

## Authentication

### Current: Environment-Based

The Lambda function authenticates to AI services using environment variables set during deployment.

**No client authentication required by default.**

### Production: API Key (Recommended)

For production, enable API Gateway API key authentication:

**1. Add to your requests:**
```bash
curl -H "X-API-Key: your-api-key" https://YOUR_API_URL/generate ...
```

**2. TypeScript example:**
```typescript
const headers = {
  'Content-Type': 'application/json',
  'X-API-Key': process.env.GIMAGE_API_KEY
};
```

**3. Configure in CDK:**
```typescript
const apiKey = api.addApiKey('ApiKey');
const plan = api.addUsagePlan('UsagePlan', {
  throttle: { rateLimit: 100, burstLimit: 200 }
});
plan.addApiKey(apiKey);
```

---

## Client SDKs

### TypeScript/JavaScript

#### Installation

```bash
npm install axios
# or
npm install @aws-sdk/client-s3
```

#### Client Class

```typescript
// gimage-client.ts
import axios, { AxiosInstance } from 'axios';

export interface GenerateImageOptions {
  prompt: string;
  model?: string;
  size?: string;
  style?: 'photorealistic' | 'artistic' | 'anime';
  negative_prompt?: string;
  seed?: number;
  response_format?: 'base64' | 's3_url';
}

export interface ImageResponse {
  image?: string;        // base64
  s3_url?: string;       // presigned URL
  s3_key?: string;
  width: number;
  height: number;
  format: string;
  size_bytes: number;
}

export interface ResizeOptions {
  image: string;         // base64 or S3 key
  width: number;
  height: number;
  response_format?: 'base64' | 's3_url';
}

export class GimageClient {
  private client: AxiosInstance;

  constructor(baseURL: string, apiKey?: string) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
        ...(apiKey && { 'X-API-Key': apiKey }),
      },
      timeout: 60000, // 60 seconds
    });
  }

  async generateImage(options: GenerateImageOptions): Promise<ImageResponse> {
    const response = await this.client.post('/generate', options);
    return response.data;
  }

  async resizeImage(options: ResizeOptions): Promise<ImageResponse> {
    const response = await this.client.post('/resize', options);
    return response.data;
  }

  async scaleImage(
    image: string,
    factor: number,
    responseFormat?: 'base64' | 's3_url'
  ): Promise<ImageResponse> {
    const response = await this.client.post('/scale', {
      image,
      factor,
      response_format: responseFormat,
    });
    return response.data;
  }

  async cropImage(
    image: string,
    x: number,
    y: number,
    width: number,
    height: number,
    responseFormat?: 'base64' | 's3_url'
  ): Promise<ImageResponse> {
    const response = await this.client.post('/crop', {
      image,
      x,
      y,
      width,
      height,
      response_format: responseFormat,
    });
    return response.data;
  }

  async compressImage(
    image: string,
    quality: number = 85,
    format?: string,
    responseFormat?: 'base64' | 's3_url'
  ): Promise<ImageResponse> {
    const response = await this.client.post('/compress', {
      image,
      quality,
      format,
      response_format: responseFormat,
    });
    return response.data;
  }

  async convertImage(
    image: string,
    targetFormat: string,
    responseFormat?: 'base64' | 's3_url'
  ): Promise<ImageResponse> {
    const response = await this.client.post('/convert', {
      image,
      target_format: targetFormat,
      response_format: responseFormat,
    });
    return response.data;
  }

  async batchProcess(operations: any[]): Promise<any> {
    const response = await this.client.post('/batch', {
      operations,
    });
    return response.data;
  }

  async healthCheck(): Promise<any> {
    const response = await this.client.get('/health');
    return response.data;
  }
}

// Usage
const client = new GimageClient(
  'https://your-api-id.execute-api.us-east-1.amazonaws.com/prod',
  process.env.GIMAGE_API_KEY
);

// Generate image
const result = await client.generateImage({
  prompt: 'a beautiful sunset',
  size: '1024x1024',
  response_format: 's3_url',
});

console.log('Image URL:', result.s3_url);
```

#### React Hook

```typescript
// useGimage.ts
import { useState, useCallback } from 'react';
import { GimageClient, GenerateImageOptions, ImageResponse } from './gimage-client';

export function useGimage(apiUrl: string, apiKey?: string) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [result, setResult] = useState<ImageResponse | null>(null);

  const client = useMemo(
    () => new GimageClient(apiUrl, apiKey),
    [apiUrl, apiKey]
  );

  const generateImage = useCallback(
    async (options: GenerateImageOptions) => {
      setLoading(true);
      setError(null);
      try {
        const data = await client.generateImage(options);
        setResult(data);
        return data;
      } catch (err) {
        const error = err as Error;
        setError(error);
        throw error;
      } finally {
        setLoading(false);
      }
    },
    [client]
  );

  const resizeImage = useCallback(
    async (file: File, width: number, height: number) => {
      setLoading(true);
      setError(null);
      try {
        const base64 = await fileToBase64(file);
        const data = await client.resizeImage({
          image: base64,
          width,
          height,
          response_format: 'base64',
        });
        setResult(data);
        return data;
      } catch (err) {
        const error = err as Error;
        setError(error);
        throw error;
      } finally {
        setLoading(false);
      }
    },
    [client]
  );

  return {
    generateImage,
    resizeImage,
    loading,
    error,
    result,
  };
}

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = reader.result as string;
      // Remove data URL prefix
      const base64 = result.split(',')[1];
      resolve(base64);
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}
```

#### React Component Example

```tsx
// ImageGenerator.tsx
import { useState } from 'react';
import { useGimage } from './useGimage';

export function ImageGenerator() {
  const [prompt, setPrompt] = useState('');
  const { generateImage, loading, error, result } = useGimage(
    process.env.REACT_APP_GIMAGE_API_URL!
  );

  const handleGenerate = async () => {
    await generateImage({
      prompt,
      size: '1024x1024',
      response_format: 's3_url',
    });
  };

  return (
    <div className="p-4">
      <h2 className="text-2xl font-bold mb-4">AI Image Generator</h2>

      <div className="mb-4">
        <input
          type="text"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder="Describe your image..."
          className="w-full p-2 border rounded"
        />
      </div>

      <button
        onClick={handleGenerate}
        disabled={loading || !prompt}
        className="bg-blue-500 text-white px-4 py-2 rounded disabled:opacity-50"
      >
        {loading ? 'Generating...' : 'Generate Image'}
      </button>

      {error && (
        <div className="mt-4 p-4 bg-red-100 text-red-700 rounded">
          Error: {error.message}
        </div>
      )}

      {result?.s3_url && (
        <div className="mt-4">
          <img
            src={result.s3_url}
            alt="Generated"
            className="max-w-full rounded shadow-lg"
          />
          <p className="mt-2 text-sm text-gray-600">
            {result.width} × {result.height} | {(result.size_bytes / 1024).toFixed(1)} KB
          </p>
        </div>
      )}
    </div>
  );
}
```

### Python

#### Installation

```bash
pip install requests pillow
```

#### Client Class

```python
# gimage_client.py
import requests
import base64
from typing import Optional, Dict, Any, List
from io import BytesIO
from PIL import Image


class GimageClient:
    def __init__(self, base_url: str, api_key: Optional[str] = None):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})

        if api_key:
            self.session.headers.update({'X-API-Key': api_key})

    def generate_image(
        self,
        prompt: str,
        model: Optional[str] = None,
        size: str = '1024x1024',
        style: Optional[str] = None,
        negative_prompt: Optional[str] = None,
        seed: Optional[int] = None,
        response_format: str = 's3_url'
    ) -> Dict[str, Any]:
        """Generate an image from a text prompt."""
        payload = {
            'prompt': prompt,
            'size': size,
            'response_format': response_format,
        }

        if model:
            payload['model'] = model
        if style:
            payload['style'] = style
        if negative_prompt:
            payload['negative_prompt'] = negative_prompt
        if seed is not None:
            payload['seed'] = seed

        response = self.session.post(f'{self.base_url}/generate', json=payload)
        response.raise_for_status()
        return response.json()

    def resize_image(
        self,
        image: str,
        width: int,
        height: int,
        response_format: str = 'base64'
    ) -> Dict[str, Any]:
        """Resize an image to specific dimensions."""
        payload = {
            'image': image,
            'width': width,
            'height': height,
            'response_format': response_format,
        }

        response = self.session.post(f'{self.base_url}/resize', json=payload)
        response.raise_for_status()
        return response.json()

    def scale_image(
        self,
        image: str,
        factor: float,
        response_format: str = 'base64'
    ) -> Dict[str, Any]:
        """Scale an image by a factor."""
        payload = {
            'image': image,
            'factor': factor,
            'response_format': response_format,
        }

        response = self.session.post(f'{self.base_url}/scale', json=payload)
        response.raise_for_status()
        return response.json()

    def crop_image(
        self,
        image: str,
        x: int,
        y: int,
        width: int,
        height: int,
        response_format: str = 'base64'
    ) -> Dict[str, Any]:
        """Crop an image to a specific region."""
        payload = {
            'image': image,
            'x': x,
            'y': y,
            'width': width,
            'height': height,
            'response_format': response_format,
        }

        response = self.session.post(f'{self.base_url}/crop', json=payload)
        response.raise_for_status()
        return response.json()

    def compress_image(
        self,
        image: str,
        quality: int = 85,
        format: Optional[str] = None,
        response_format: str = 'base64'
    ) -> Dict[str, Any]:
        """Compress an image with quality settings."""
        payload = {
            'image': image,
            'quality': quality,
            'response_format': response_format,
        }

        if format:
            payload['format'] = format

        response = self.session.post(f'{self.base_url}/compress', json=payload)
        response.raise_for_status()
        return response.json()

    def convert_image(
        self,
        image: str,
        target_format: str,
        response_format: str = 'base64'
    ) -> Dict[str, Any]:
        """Convert an image to a different format."""
        payload = {
            'image': image,
            'target_format': target_format,
            'response_format': response_format,
        }

        response = self.session.post(f'{self.base_url}/convert', json=payload)
        response.raise_for_status()
        return response.json()

    def batch_process(self, operations: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Process multiple images in batch."""
        payload = {'operations': operations}

        response = self.session.post(f'{self.base_url}/batch', json=payload)
        response.raise_for_status()
        return response.json()

    def health_check(self) -> Dict[str, Any]:
        """Check API health."""
        response = self.session.get(f'{self.base_url}/health')
        response.raise_for_status()
        return response.json()

    # Helper methods

    @staticmethod
    def image_to_base64(image_path: str) -> str:
        """Convert image file to base64 string."""
        with open(image_path, 'rb') as f:
            return base64.b64encode(f.read()).decode('utf-8')

    @staticmethod
    def base64_to_image(base64_str: str) -> Image.Image:
        """Convert base64 string to PIL Image."""
        image_data = base64.b64decode(base64_str)
        return Image.open(BytesIO(image_data))

    @staticmethod
    def save_base64_image(base64_str: str, output_path: str):
        """Save base64 image to file."""
        img = GimageClient.base64_to_image(base64_str)
        img.save(output_path)


# Usage example
if __name__ == '__main__':
    import os

    client = GimageClient(
        base_url='https://your-api-id.execute-api.us-east-1.amazonaws.com/prod',
        api_key=os.getenv('GIMAGE_API_KEY')
    )

    # Generate image
    result = client.generate_image(
        prompt='a beautiful sunset over mountains',
        size='1024x1024',
        response_format='s3_url'
    )

    print(f"Generated image: {result['s3_url']}")
    print(f"Dimensions: {result['width']}x{result['height']}")

    # Resize image from file
    base64_image = client.image_to_base64('photo.jpg')
    resized = client.resize_image(
        image=base64_image,
        width=800,
        height=600,
        response_format='base64'
    )

    # Save result
    client.save_base64_image(resized['image'], 'resized.jpg')
    print(f"Resized image saved: {resized['width']}x{resized['height']}")
```

#### Flask Integration Example

```python
# app.py
from flask import Flask, request, jsonify, send_file
from gimage_client import GimageClient
import os
from io import BytesIO

app = Flask(__name__)
client = GimageClient(
    base_url=os.getenv('GIMAGE_API_URL'),
    api_key=os.getenv('GIMAGE_API_KEY')
)

@app.route('/api/generate', methods=['POST'])
def generate():
    data = request.json
    prompt = data.get('prompt')

    if not prompt:
        return jsonify({'error': 'Prompt is required'}), 400

    try:
        result = client.generate_image(
            prompt=prompt,
            size=data.get('size', '1024x1024'),
            response_format='s3_url'
        )
        return jsonify(result)
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/resize', methods=['POST'])
def resize():
    if 'file' not in request.files:
        return jsonify({'error': 'File is required'}), 400

    file = request.files['file']
    width = int(request.form.get('width', 800))
    height = int(request.form.get('height', 600))

    # Convert to base64
    base64_image = GimageClient.image_to_base64(file.stream)

    try:
        result = client.resize_image(
            image=base64_image,
            width=width,
            height=height,
            response_format='base64'
        )

        # Convert back to image
        img = GimageClient.base64_to_image(result['image'])

        # Return as file
        img_io = BytesIO()
        img.save(img_io, 'PNG')
        img_io.seek(0)

        return send_file(img_io, mimetype='image/png')
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True)
```

### Go

```go
// gimage_client.go
package gimage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type GenerateRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Seed           int64  `json:"seed,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

type ImageResponse struct {
	Image     string `json:"image,omitempty"`
	S3URL     string `json:"s3_url,omitempty"`
	S3Key     string `json:"s3_key,omitempty"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	SizeBytes int64  `json:"size_bytes"`
}

type ResizeRequest struct {
	Image          string `json:"image"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	ResponseFormat string `json:"response_format,omitempty"`
}

func NewClient(baseURL string, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, errResp.Error, errResp.Message)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) GenerateImage(req GenerateRequest) (*ImageResponse, error) {
	var result ImageResponse
	err := c.doRequest("POST", "/generate", req, &result)
	return &result, err
}

func (c *Client) ResizeImage(req ResizeRequest) (*ImageResponse, error) {
	var result ImageResponse
	err := c.doRequest("POST", "/resize", req, &result)
	return &result, err
}

func (c *Client) HealthCheck() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest("GET", "/health", nil, &result)
	return result, err
}

// Helper: Convert file to base64
func FileToBase64(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// Helper: Save base64 to file
func Base64ToFile(base64Str string, outputPath string) error {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

// Usage example
func main() {
	client := NewClient(
		"https://your-api-id.execute-api.us-east-1.amazonaws.com/prod",
		os.Getenv("GIMAGE_API_KEY"),
	)

	// Generate image
	result, err := client.GenerateImage(GenerateRequest{
		Prompt:         "a beautiful sunset",
		Size:           "1024x1024",
		ResponseFormat: "s3_url",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated: %s (%dx%d)\n", result.S3URL, result.Width, result.Height)
}
```

### cURL

#### Generate Image

```bash
curl -X POST https://YOUR_API_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "a sunset over mountains",
    "size": "1024x1024",
    "response_format": "s3_url"
  }'
```

#### Resize Image (Base64)

```bash
# First, convert image to base64
BASE64_IMAGE=$(base64 -i photo.jpg)

curl -X POST https://YOUR_API_URL/resize \
  -H "Content-Type: application/json" \
  -d "{
    \"image\": \"$BASE64_IMAGE\",
    \"width\": 800,
    \"height\": 600,
    \"response_format\": \"base64\"
  }"
```

#### Batch Processing

```bash
curl -X POST https://YOUR_API_URL/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "operation": "resize",
        "image": "images/photo1.jpg",
        "params": { "width": 800, "height": 600 }
      },
      {
        "operation": "compress",
        "image": "images/photo2.png",
        "params": { "quality": 85, "format": "webp" }
      }
    ]
  }'
```

---

## Common Integration Patterns

### Pattern 1: User Upload → Process → Display

```typescript
async function handleImageUpload(file: File) {
  // 1. Convert to base64
  const base64 = await fileToBase64(file);

  // 2. Resize for web
  const resized = await client.resizeImage({
    image: base64,
    width: 1200,
    height: 800,
    response_format: 'base64'
  });

  // 3. Display in UI
  return `data:image/${resized.format};base64,${resized.image}`;
}
```

### Pattern 2: AI Generation → Save to Storage

```typescript
async function generateAndStore(prompt: string) {
  // 1. Generate with S3 response
  const generated = await client.generateImage({
    prompt,
    size: '1024x1024',
    response_format: 's3_url'
  });

  // 2. Download from S3 presigned URL
  const imageData = await fetch(generated.s3_url);
  const blob = await imageData.blob();

  // 3. Upload to your own storage
  await uploadToYourStorage(blob, `generated-${Date.now()}.png`);

  return generated.s3_key;
}
```

### Pattern 3: Pipeline Processing

```typescript
async function processImagePipeline(imageUrl: string) {
  // 1. Download original
  const response = await fetch(imageUrl);
  const blob = await response.blob();
  const base64 = await blobToBase64(blob);

  // 2. Resize
  const resized = await client.resizeImage({
    image: base64,
    width: 1920,
    height: 1080,
    response_format: 'base64'
  });

  // 3. Compress
  const compressed = await client.compressImage(
    resized.image,
    85,
    'webp',
    'base64'
  );

  // 4. Return final result
  return compressed;
}
```

### Pattern 4: Batch Thumbnail Generation

```typescript
async function generateThumbnails(imageKeys: string[]) {
  const operations = imageKeys.map(key => ({
    operation: 'resize',
    image: key,
    params: { width: 200, height: 200 }
  }));

  const result = await client.batchProcess(operations);

  return result.results.map((r, i) => ({
    original: imageKeys[i],
    thumbnail: r.s3_url
  }));
}
```

---

## Error Handling

### Error Response Format

```json
{
  "error": "Bad Request",
  "message": "Width must be positive",
  "code": 400
}
```

### TypeScript Error Handling

```typescript
try {
  const result = await client.generateImage(options);
  return result;
} catch (error) {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    const data = error.response?.data;

    if (status === 400) {
      console.error('Invalid request:', data.message);
    } else if (status === 500) {
      console.error('Server error:', data.message);
    } else if (status === 503) {
      console.error('Service unavailable');
    }

    throw new Error(`API Error: ${data.message}`);
  }
  throw error;
}
```

### Python Error Handling

```python
from requests.exceptions import HTTPError, Timeout

try:
    result = client.generate_image(prompt='test')
except HTTPError as e:
    if e.response.status_code == 400:
        print(f"Bad request: {e.response.json()['message']}")
    elif e.response.status_code == 500:
        print(f"Server error: {e.response.json()['message']}")
    else:
        print(f"HTTP error: {e}")
except Timeout:
    print("Request timed out")
except Exception as e:
    print(f"Unexpected error: {e}")
```

---

## Best Practices

### 1. Use Appropriate Response Formats

```typescript
// For small images (thumbnails, icons)
response_format: 'base64'  // Faster, embedded in response

// For large images (full resolution)
response_format: 's3_url'   // Efficient, uses presigned URLs
```

### 2. Implement Retry Logic

```typescript
async function retryRequest<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3
): Promise<T> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === maxRetries - 1) throw error;

      // Exponential backoff
      const delay = Math.pow(2, i) * 1000;
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  throw new Error('Max retries exceeded');
}

// Usage
const result = await retryRequest(() =>
  client.generateImage({ prompt: 'test' })
);
```

### 3. Cache S3 URLs

```typescript
const cache = new Map<string, { url: string; expires: number }>();

function getCachedS3URL(s3Key: string): string | null {
  const cached = cache.get(s3Key);
  if (cached && cached.expires > Date.now()) {
    return cached.url;
  }
  return null;
}

function cacheS3URL(s3Key: string, url: string) {
  // URLs expire in 60 minutes
  cache.set(s3Key, {
    url,
    expires: Date.now() + 59 * 60 * 1000
  });
}
```

### 4. Validate Input Dimensions

```typescript
function validateDimensions(width: number, height: number) {
  if (width <= 0 || height <= 0) {
    throw new Error('Dimensions must be positive');
  }
  if (width > 10000 || height > 10000) {
    throw new Error('Dimensions too large (max 10000)');
  }
}
```

### 5. Monitor Usage

```typescript
let requestCount = 0;
let totalBytes = 0;

client.interceptors.response.use(response => {
  requestCount++;
  if (response.data.size_bytes) {
    totalBytes += response.data.size_bytes;
  }

  console.log(`Requests: ${requestCount}, Total data: ${(totalBytes / 1024 / 1024).toFixed(2)} MB`);

  return response;
});
```

---

## Rate Limits & Quotas

### Default Limits (Configure in API Gateway)

- **Rate**: 100 requests/second
- **Burst**: 200 requests
- **Daily quota**: 10,000 requests

### Handling Rate Limits

```typescript
function isRateLimitError(error: any): boolean {
  return error.response?.status === 429;
}

async function handleRateLimit<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn();
  } catch (error) {
    if (isRateLimitError(error)) {
      // Wait and retry
      await new Promise(resolve => setTimeout(resolve, 1000));
      return fn();
    }
    throw error;
  }
}
```

---

## Examples by Use Case

### Web Application: Image Gallery

```typescript
class ImageGallery {
  constructor(private client: GimageClient) {}

  async uploadAndProcess(file: File): Promise<GalleryImage> {
    // 1. Convert to base64
    const base64 = await fileToBase64(file);

    // 2. Create thumbnail
    const thumbnail = await this.client.resizeImage({
      image: base64,
      width: 200,
      height: 200,
      response_format: 'base64'
    });

    // 3. Optimize original
    const optimized = await this.client.compressImage(
      base64,
      85,
      'webp',
      's3_url'
    );

    return {
      thumbnail: `data:image/${thumbnail.format};base64,${thumbnail.image}`,
      full: optimized.s3_url,
      width: optimized.width,
      height: optimized.height
    };
  }
}
```

### E-commerce: Product Images

```python
class ProductImageProcessor:
    def __init__(self, client: GimageClient):
        self.client = client

    def process_product_image(self, image_path: str) -> dict:
        """Process product image into multiple sizes."""
        base64_image = self.client.image_to_base64(image_path)

        # Batch process multiple sizes
        operations = [
            {
                'operation': 'resize',
                'image': base64_image,
                'params': {'width': 1200, 'height': 1200}  # Large
            },
            {
                'operation': 'resize',
                'image': base64_image,
                'params': {'width': 600, 'height': 600}    # Medium
            },
            {
                'operation': 'resize',
                'image': base64_image,
                'params': {'width': 200, 'height': 200}    # Thumbnail
            }
        ]

        result = self.client.batch_process(operations)

        return {
            'large': result['results'][0]['s3_url'],
            'medium': result['results'][1]['s3_url'],
            'thumbnail': result['results'][2]['s3_url']
        }
```

### Social Media: Auto-Generate Images

```go
func GenerateSocialMediaImage(client *gimage.Client, text string) (string, error) {
	// 1. Generate image from text
	result, err := client.GenerateImage(gimage.GenerateRequest{
		Prompt:         fmt.Sprintf("Social media post: %s", text),
		Size:           "1200x630", // Open Graph size
		Style:          "artistic",
		ResponseFormat: "s3_url",
	})
	if err != nil {
		return "", err
	}

	// 2. Add text overlay (separate service)
	// ...

	return result.S3URL, nil
}
```

---

## Next Steps

1. **Review OpenAPI Spec**: See `openapi.yaml` for complete API reference
2. **Try Examples**: Use provided client SDKs in your project
3. **Set Up Monitoring**: Track usage and performance
4. **Enable Authentication**: Add API key for production
5. **Optimize Costs**: Use batch processing and caching

## Support

- **API Reference**: `openapi.yaml`
- **GitHub Issues**: https://github.com/chadneal/gimage/issues
- **Documentation**: `lambda.md`, `LAMBDA_STATUS.md`

---

**Built with ❤️ for developers integrating AI-powered image processing**
