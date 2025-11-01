# Gimage Lambda API - Quick Reference

Quick reference guide for the Gimage Lambda API endpoints.

## Base URL

```
https://YOUR_API_ID.execute-api.REGION.amazonaws.com/prod
```

## Runtime Information

- **Platform**: AWS Lambda
- **Runtime**: provided.al2023 (Amazon Linux 2023)
- **Architecture**: ARM64 (Graviton2)
- **Language**: Go 1.22+
- **Memory**: 2048 MB
- **Timeout**: 5 minutes

---

## Endpoints

### POST /generate

Generate AI images from text prompts.

**Request:**
```json
{
  "prompt": "a sunset over mountains",
  "model": "gemini-2.5-flash-image",
  "size": "1024x1024",
  "style": "photorealistic",
  "negative_prompt": "people, text",
  "seed": 42,
  "response_format": "s3_url"
}
```

**Response (base64):**
```json
{
  "image": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB...",
  "width": 1024,
  "height": 1024,
  "format": "png",
  "size_bytes": 524288
}
```

**Response (S3):**
```json
{
  "s3_url": "https://bucket.s3.amazonaws.com/key?presigned...",
  "s3_key": "images/1234567890-abc.png",
  "width": 2048,
  "height": 2048,
  "format": "png",
  "size_bytes": 2097152
}
```

**Models:**
- `gemini-2.5-flash-image` (default)
- `gemini-2.0-flash-preview-image-generation`
- `imagen-3.0-generate-001`
- `imagen-4.0-generate-001`
- `imagen-4.0-ultra-generate-001`
- `imagen-4.0-fast-generate-001`

---

### POST /resize

Resize image to specific dimensions.

**Request:**
```json
{
  "image": "base64_or_s3_key",
  "width": 800,
  "height": 600,
  "response_format": "base64"
}
```

**Response:** Same as generate endpoint

---

### POST /scale

Scale image by factor.

**Request:**
```json
{
  "image": "base64_or_s3_key",
  "factor": 0.5,
  "response_format": "base64"
}
```

**Response:** Same as generate endpoint

---

### POST /crop

Crop image to region.

**Request:**
```json
{
  "image": "base64_or_s3_key",
  "x": 100,
  "y": 100,
  "width": 800,
  "height": 600,
  "response_format": "base64"
}
```

**Response:** Same as generate endpoint

---

### POST /compress

Compress image with quality settings.

**Request:**
```json
{
  "image": "base64_or_s3_key",
  "quality": 85,
  "format": "jpg",
  "response_format": "base64"
}
```

**Quality:** 1-100 (higher = better quality, larger file)

**Response:** Same as generate endpoint

---

### POST /convert

Convert image format.

**Request:**
```json
{
  "image": "base64_or_s3_key",
  "target_format": "webp",
  "response_format": "base64"
}
```

**Formats:**
- `png` - Portable Network Graphics
- `jpg`, `jpeg` - JPEG
- `gif` - Graphics Interchange Format
- `webp` - WebP
- `tiff`, `tif` - TIFF
- `bmp` - Bitmap

**Response:** Same as generate endpoint

---

### POST /batch

Process multiple images concurrently.

**Request:**
```json
{
  "operations": [
    {
      "operation": "resize",
      "image": "s3_key_or_base64",
      "params": {
        "width": 800,
        "height": 600
      }
    },
    {
      "operation": "compress",
      "image": "s3_key_or_base64",
      "params": {
        "quality": 85,
        "format": "webp"
      }
    }
  ]
}
```

**Operations:** `resize`, `scale`, `crop`, `compress`, `convert`

**Response:**
```json
{
  "batch_id": "batch-1234567890",
  "status": "completed",
  "results": [
    {
      "s3_url": "https://...",
      "width": 800,
      "height": 600,
      "format": "png",
      "size_bytes": 102400
    },
    {
      "s3_url": "https://...",
      "width": 1024,
      "height": 768,
      "format": "webp",
      "size_bytes": 51200
    }
  ]
}
```

---

### GET /health

Health check endpoint.

**Response:**
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

---

## Error Responses

All errors follow this format:

```json
{
  "error": "Bad Request",
  "message": "Width must be positive",
  "code": 400
}
```

**Common Status Codes:**
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `404` - Not Found (invalid endpoint)
- `500` - Internal Server Error
- `503` - Service Unavailable

---

## Request/Response Formats

### Image Input

Two formats supported:

**1. Base64 (recommended for small images < 5MB):**
```json
{
  "image": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB..."
}
```

**2. S3 Key (from previous operation):**
```json
{
  "image": "images/1234567890-abc.png"
}
```

### Response Format

Controlled by `response_format` parameter:

**base64** (for images < 512KB):
```json
{
  "image": "base64_encoded_data",
  "width": 800,
  "height": 600,
  "format": "png",
  "size_bytes": 102400
}
```

**s3_url** (for larger images):
```json
{
  "s3_url": "https://presigned-url-valid-60-min",
  "s3_key": "images/key.png",
  "width": 2048,
  "height": 2048,
  "format": "png",
  "size_bytes": 2097152
}
```

---

## Headers

**Request:**
```
Content-Type: application/json
X-API-Key: your-api-key (if authentication enabled)
```

**Response:**
```
Content-Type: application/json
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-API-Key
```

---

## Rate Limits

Default limits (configurable in API Gateway):

- **Rate**: 100 requests/second
- **Burst**: 200 requests
- **Daily**: 10,000 requests

---

## Examples

### cURL

**Generate Image:**
```bash
curl -X POST https://YOUR_API_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "beautiful sunset",
    "size": "1024x1024",
    "response_format": "s3_url"
  }'
```

**Resize Image:**
```bash
BASE64=$(base64 -i photo.jpg)
curl -X POST https://YOUR_API_URL/resize \
  -H "Content-Type: application/json" \
  -d "{
    \"image\": \"$BASE64\",
    \"width\": 800,
    \"height\": 600
  }"
```

### JavaScript

```javascript
// Generate
const response = await fetch('https://YOUR_API_URL/generate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    prompt: 'sunset',
    size: '1024x1024',
    response_format: 's3_url'
  })
});

const result = await response.json();
console.log(result.s3_url);

// Resize
const resizeResponse = await fetch('https://YOUR_API_URL/resize', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    image: base64Image,
    width: 800,
    height: 600,
    response_format: 'base64'
  })
});

const resized = await resizeResponse.json();
// Use resized.image
```

### Python

```python
import requests

# Generate
response = requests.post('https://YOUR_API_URL/generate', json={
    'prompt': 'beautiful landscape',
    'size': '1024x1024',
    'response_format': 's3_url'
})

result = response.json()
print(result['s3_url'])

# Resize
response = requests.post('https://YOUR_API_URL/resize', json={
    'image': base64_image,
    'width': 800,
    'height': 600
})

result = response.json()
print(f"{result['width']}x{result['height']}")
```

---

## Cost Estimate

For 10,000 monthly requests:

| Service | Cost |
|---------|------|
| Lambda (2GB, ARM64) | $0.17 |
| Lambda Requests | $0.002 |
| S3 Storage (1GB) | $0.023 |
| S3 Requests | $0.01 |
| API Gateway | $0.035 |
| CloudWatch Logs | $0.01 |
| **Total** | **~$0.25/month** |

Plus Gemini/Vertex AI costs (separate).

---

## Environment Variables

Lambda function requires these environment variables:

**Required:**
- `S3_BUCKET` - S3 bucket name
- `AWS_REGION` - AWS region
- `GEMINI_API_KEY` - Gemini API key

**Optional:**
- `VERTEX_API_KEY` - Vertex AI Express Mode key
- `VERTEX_PROJECT` - GCP project ID
- `VERTEX_LOCATION` - Vertex location (default: us-central1)
- `MAX_RESPONSE_SIZE_KB` - Max size for base64 (default: 512)
- `PRESIGNED_URL_EXPIRATION_MINUTES` - S3 URL expiration (default: 60)
- `LOG_LEVEL` - Logging level (default: info)

---

## Documentation

- **OpenAPI Spec**: [openapi.yaml](openapi.yaml)
- **Integration Guide**: [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)
- **Deployment Guide**: [lambda.md](lambda.md)
- **Status**: [LAMBDA_STATUS.md](LAMBDA_STATUS.md)

---

## Support

- GitHub Issues: https://github.com/apresai/gimage/issues
- Discussions: https://github.com/apresai/gimage/discussions
