# Gimage AWS Lambda Deployment

Deploy gimage as a serverless REST API on AWS Lambda for web application integration.

## Features

- **Serverless**: Auto-scales from 0 to thousands of requests
- **ARM64/Graviton2**: Fast performance, low cost
- **S3 Integration**: Automatic image storage and CDN delivery
- **Production Ready**: Full CORS, error handling, monitoring

## Quick Start

### Prerequisites

- AWS Account with CLI configured
- Node.js 20+ (for CDK)
- Go 1.22+ (for building)

### Deploy in 3 Steps

```bash
# 1. Build the Lambda function
make build-lambda

# 2. Package for deployment
make package-lambda

# 3. Deploy infrastructure with CDK
cd infrastructure/cdk
npm install
cdk deploy
```

### Configure API Credentials

Set environment variables in Lambda console or CDK:

```bash
# For Gemini API
GEMINI_API_KEY=your-gemini-key

# For Vertex AI
VERTEX_API_KEY=your-vertex-key
VERTEX_PROJECT=your-gcp-project
VERTEX_LOCATION=us-central1

# For AWS Bedrock
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=wJalr...
AWS_REGION=us-east-1
```

## API Endpoints

Base URL: `https://YOUR_API_ID.execute-api.REGION.amazonaws.com/prod`

### Image Generation

**POST** `/generate`

```json
{
  "prompt": "a sunset over mountains",
  "model": "gemini-2.5-flash-image",
  "size": "1024x1024",
  "style": "photorealistic",
  "response_format": "s3_url"
}
```

**Response:**
```json
{
  "s3_url": "https://s3.amazonaws.com/...",
  "width": 1024,
  "height": 1024,
  "format": "png"
}
```

### Image Processing

**POST** `/resize`
```json
{
  "image": "base64_or_s3_key",
  "width": 800,
  "height": 600
}
```

**POST** `/scale`
```json
{
  "image": "base64_or_s3_key",
  "factor": 0.5
}
```

**POST** `/crop`
```json
{
  "image": "base64_or_s3_key",
  "x": 100,
  "y": 100,
  "width": 800,
  "height": 600
}
```

**POST** `/compress`
```json
{
  "image": "base64_or_s3_key",
  "quality": 85,
  "format": "webp"
}
```

**POST** `/convert`
```json
{
  "image": "base64_or_s3_key",
  "target_format": "webp"
}
```

### Batch Processing

**POST** `/batch`
```json
{
  "operations": [
    {
      "operation": "resize",
      "image": "s3_key",
      "width": 800,
      "height": 600
    }
  ]
}
```

### Health Check

**GET** `/health`
```json
{
  "status": "healthy",
  "version": "0.1.1"
}
```

## Configuration

### Environment Variables

Set in Lambda function configuration:

| Variable | Required | Description |
|----------|----------|-------------|
| `GEMINI_API_KEY` | Optional | Gemini API key |
| `VERTEX_API_KEY` | Optional | Vertex AI API key |
| `VERTEX_PROJECT` | Optional | GCP project ID |
| `VERTEX_LOCATION` | Optional | Vertex AI location |
| `AWS_REGION` | Optional | AWS region for Bedrock |
| `S3_BUCKET` | Auto-set | S3 bucket for images |

### Response Formats

The API automatically chooses response format based on image size:

- **Small images (<512KB)**: Base64 in JSON response
- **Large images (>512KB)**: S3 presigned URL (24h expiry)

You can force a format with `response_format: "base64"` or `"s3_url"`

## Architecture

```
Client → API Gateway → Lambda (ARM64/Go) → S3 + AI APIs
                                              ↓
                                    Presigned URLs / Base64
```

**Runtime**: provided.al2023 (Amazon Linux 2023)
**Architecture**: ARM64 (Graviton2)
**Memory**: 2048 MB
**Timeout**: 5 minutes
**Package Size**: ~17MB compressed

## Monitoring

CloudWatch logs automatically capture:
- Request/response details
- Error traces
- Performance metrics
- API usage

Access logs at: CloudWatch → Log Groups → `/aws/lambda/gimage-lambda`

## Cost Estimates

**For 10,000 requests/month:**
- Lambda: ~$0.25/month
- API Gateway: ~$0.35/month
- S3: ~$0.05/month
- **Total: ~$0.65/month**

Plus AI API costs (Gemini free tier: 1500/day)

## Troubleshooting

### Lambda function fails to start
- Check CloudWatch logs
- Verify environment variables are set
- Ensure Lambda has IAM permissions for S3

### API returns 403 errors
- Verify API Gateway CORS configuration
- Check API key requirements
- Review IAM execution role permissions

### Images not uploading to S3
- Verify S3 bucket exists
- Check Lambda IAM role has S3 permissions
- Review S3_BUCKET environment variable

### Generation API returns errors
- Verify API credentials (Gemini/Vertex/Bedrock)
- Check quota limits haven't been exceeded
- Review CloudWatch logs for detailed errors

## Development

### Local Testing

```bash
# Build function
make build-lambda

# Test locally (requires SAM CLI)
sam local invoke -e test-event.json

# Or deploy to test environment
cdk deploy --profile dev
```

### Update Deployment

```bash
# Rebuild and redeploy
make build-lambda && make package-lambda
cd infrastructure/cdk && cdk deploy
```

## Documentation

- **Integration Guide**: [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) - Client examples
- **OpenAPI Spec**: [openapi.yaml](openapi.yaml) - Full API specification
- **Main README**: [README.md](README.md) - Project overview

## Next Steps

1. Deploy the Lambda function
2. Configure API credentials
3. Test with `/health` endpoint
4. Generate your first image
5. Integrate into your application

See [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) for client code examples in TypeScript, Python, and Go.
