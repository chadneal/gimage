# Gimage Lambda API - Integration Guide

Quick integration examples for the Gimage Lambda API.

## Quick Start

### Base URL

After deployment, your API will be available at:
```
https://YOUR_API_ID.execute-api.REGION.amazonaws.com/prod
```

### Health Check

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

## Authentication

Currently API Gateway handles authentication. Configure in AWS Console:
- API Key required (optional)
- IAM authorization (optional)
- Cognito (optional)

## Response Formats

All endpoints support two response modes:

1. **Base64** (images <512KB) - Image data in JSON response
2. **S3 URL** (images >512KB) - Presigned URL (24h expiry)

Control with `response_format` parameter: `"base64"` or `"s3_url"`

## Client Examples

### TypeScript/JavaScript

```typescript
// Generate image
async function generateImage() {
  const response = await fetch('https://YOUR_API_URL/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prompt: 'a sunset over mountains',
      size: '1024x1024',
      response_format: 's3_url'
    })
  });

  const result = await response.json();
  console.log('Image URL:', result.s3_url);
  return result;
}

// Resize image
async function resizeImage(imageData: string) {
  const response = await fetch('https://YOUR_API_URL/resize', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      image: imageData, // base64 or s3_key
      width: 800,
      height: 600,
      response_format: 'base64'
    })
  });

  const result = await response.json();
  return result.image; // base64 data
}

// Batch operations
async function batchProcess() {
  const response = await fetch('https://YOUR_API_URL/batch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      operations: [
        { operation: 'resize', image: 's3://key1', width: 800, height: 600 },
        { operation: 'compress', image: 's3://key2', quality: 85 }
      ]
    })
  });

  return await response.json();
}
```

**React Hook Example:**

```typescript
import { useState } from 'react';

export function useGimage(apiUrl: string) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const generateImage = async (prompt: string) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`${apiUrl}/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt, response_format: 's3_url' })
      });

      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }

      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { generateImage, loading, error };
}
```

### Python

```python
import requests
import base64

API_URL = "https://YOUR_API_URL"

# Generate image
def generate_image(prompt, size="1024x1024"):
    response = requests.post(
        f"{API_URL}/generate",
        json={
            "prompt": prompt,
            "size": size,
            "response_format": "s3_url"
        }
    )
    response.raise_for_status()
    return response.json()

# Resize image
def resize_image(image_path, width, height):
    # Read and encode image
    with open(image_path, 'rb') as f:
        image_data = base64.b64encode(f.read()).decode()

    response = requests.post(
        f"{API_URL}/resize",
        json={
            "image": image_data,
            "width": width,
            "height": height,
            "response_format": "base64"
        }
    )
    response.raise_for_status()

    result = response.json()

    # Decode and save result
    with open('resized.png', 'wb') as f:
        f.write(base64.b64decode(result['image']))

    return result

# Compress image
def compress_image(s3_key, quality=85):
    response = requests.post(
        f"{API_URL}/compress",
        json={
            "image": s3_key,
            "quality": quality,
            "format": "webp",
            "response_format": "s3_url"
        }
    )
    response.raise_for_status()
    return response.json()

# Example usage
if __name__ == "__main__":
    # Generate
    result = generate_image("a beautiful sunset")
    print(f"Generated: {result['s3_url']}")

    # Resize
    result = resize_image("photo.jpg", 800, 600)
    print(f"Resized: {result['width']}x{result['height']}")
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

const apiURL = "https://YOUR_API_URL"

type GenerateRequest struct {
    Prompt         string `json:"prompt"`
    Size           string `json:"size"`
    ResponseFormat string `json:"response_format"`
}

type GenerateResponse struct {
    S3URL  string `json:"s3_url"`
    Width  int    `json:"width"`
    Height int    `json:"height"`
    Format string `json:"format"`
}

func GenerateImage(prompt string) (*GenerateResponse, error) {
    reqBody, _ := json.Marshal(GenerateRequest{
        Prompt:         prompt,
        Size:           "1024x1024",
        ResponseFormat: "s3_url",
    })

    resp, err := http.Post(
        apiURL+"/generate",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }

    var result GenerateResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

type ResizeRequest struct {
    Image          string `json:"image"`
    Width          int    `json:"width"`
    Height         int    `json:"height"`
    ResponseFormat string `json:"response_format"`
}

func ResizeImage(imageData string, width, height int) (map[string]interface{}, error) {
    reqBody, _ := json.Marshal(ResizeRequest{
        Image:          imageData,
        Width:          width,
        Height:         height,
        ResponseFormat: "base64",
    })

    resp, err := http.Post(
        apiURL+"/resize",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result, nil
}

func main() {
    // Generate image
    result, err := GenerateImage("a sunset over mountains")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Generated: %s (%dx%d)\n", result.S3URL, result.Width, result.Height)
}
```

### cURL

```bash
# Generate image
curl -X POST https://YOUR_API_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "a sunset over mountains",
    "size": "1024x1024",
    "response_format": "s3_url"
  }'

# Resize image
curl -X POST https://YOUR_API_URL/resize \
  -H "Content-Type: application/json" \
  -d '{
    "image": "s3://my-bucket/image.jpg",
    "width": 800,
    "height": 600,
    "response_format": "base64"
  }'

# Compress image
curl -X POST https://YOUR_API_URL/compress \
  -H "Content-Type: application/json" \
  -d '{
    "image": "s3://my-bucket/image.jpg",
    "quality": 85,
    "format": "webp"
  }'

# Health check
curl https://YOUR_API_URL/health
```

## Error Handling

All endpoints return standard HTTP status codes:

- `200` - Success
- `400` - Bad request (invalid parameters)
- `500` - Server error
- `503` - Service unavailable

Error response format:
```json
{
  "error": "Invalid image format",
  "code": "INVALID_FORMAT",
  "details": "Expected PNG, JPG, or WebP"
}
```

### Example Error Handling (TypeScript)

```typescript
async function safeGenerateImage(prompt: string) {
  try {
    const response = await fetch('https://YOUR_API_URL/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`API Error: ${error.error}`);
    }

    return await response.json();
  } catch (err) {
    console.error('Generation failed:', err);
    throw err;
  }
}
```

## Common Patterns

### Upload Image for Processing

```typescript
// 1. Upload to S3 (if using your own bucket)
const uploadToS3 = async (file: File) => {
  // Use AWS SDK or presigned URL
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('YOUR_S3_UPLOAD_URL', {
    method: 'PUT',
    body: file
  });

  return s3Key;
};

// 2. Process with gimage API
const s3Key = await uploadToS3(file);
const result = await fetch('https://YOUR_API_URL/resize', {
  method: 'POST',
  body: JSON.stringify({
    image: s3Key,
    width: 800,
    height: 600
  })
});
```

### Generate and Download

```typescript
async function generateAndDownload(prompt: string) {
  // Generate image
  const response = await fetch('https://YOUR_API_URL/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prompt,
      response_format: 's3_url'
    })
  });

  const result = await response.json();

  // Download from S3 URL
  const imageBlob = await fetch(result.s3_url).then(r => r.blob());

  // Trigger download
  const url = URL.createObjectURL(imageBlob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'generated.png';
  a.click();
}
```

## Rate Limits

Default API Gateway limits:
- 10,000 requests per second (burst)
- 5,000 requests per second (steady state)

Lambda concurrency limits:
- 1,000 concurrent executions (default)
- Configurable per function

## Best Practices

1. **Use S3 URLs for large images** - Reduces response size and Lambda memory
2. **Handle errors gracefully** - Retry with exponential backoff
3. **Cache results** - Store generated images for reuse
4. **Optimize batch operations** - Process multiple images in one request
5. **Monitor costs** - Track API Gateway and Lambda usage

## Documentation

- **Deployment**: [lambda.md](lambda.md) - Deployment guide
- **OpenAPI Spec**: [openapi.yaml](openapi.yaml) - Full API reference
- **Main README**: [README.md](README.md) - Project overview

## Support

For issues or questions:
- GitHub Issues: https://github.com/apresai/gimage/issues
- Discussions: https://github.com/apresai/gimage/discussions
