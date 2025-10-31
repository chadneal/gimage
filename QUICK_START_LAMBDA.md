# Gimage Lambda API - Quick Start Guide

Get the Gimage Lambda API deployed in under 1 hour.

## 🎯 What You're Deploying

A production-ready serverless REST API for AI-powered image generation and processing:

- **10 API endpoints** (generate, resize, scale, crop, compress, convert, batch, health, docs, openapi)
- **Interactive Swagger UI** at `/docs` for instant testing
- **ARM64/Graviton2** for 20% cost savings
- **Auto-scaling** from 0 to thousands of requests
- **~$0.25/month** for 10K requests

## 📋 Prerequisites (5 minutes)

### Install Required Tools

```bash
# AWS CLI
brew install awscli
aws configure

# Node.js 20+
brew install node@20

# AWS CDK
npm install -g aws-cdk

# Verify installations
aws --version    # Should be 2.x+
node --version   # Should be v20.x+
cdk --version    # Should be 2.x+
```

### Get API Keys

You need **one** of these for AI image generation:

**Option A: Gemini API** (Recommended for getting started)
- Get key from: https://aistudio.google.com/app/apikey
- Free tier: 1500 requests/day

**Option B: Vertex AI**
- Setup GCP project
- Enable Vertex AI API
- Create service account or API key

## 🚀 5-Step Deployment

### Step 1: Build Lambda Package (2 minutes)

```bash
# From gimage directory
make clean-lambda
make build-lambda
make package-lambda

# Verify package
ls -lh bin/lambda.zip
# Should show ~17MB file
```

### Step 2: Create CDK Infrastructure (15 minutes)

```bash
# Create CDK directory
mkdir -p infrastructure/cdk
cd infrastructure/cdk

# Initialize CDK
cdk init app --language typescript

# Install dependencies
npm install @aws-cdk/aws-lambda \
            @aws-cdk/aws-apigateway \
            @aws-cdk/aws-s3 \
            @aws-cdk/aws-iam \
            @aws-cdk/aws-logs
```

Copy the CDK stack code from `lambda.md` to `lib/gimage-stack.ts`.

Update `bin/cdk.ts`:
```typescript
#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { GimageStack } from '../lib/gimage-stack';

const app = new cdk.App();
new GimageStack(app, 'GimageStack', {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION || 'us-east-1',
  },
});
```

### Step 3: Bootstrap CDK (First Time Only, 5 minutes)

```bash
# From infrastructure/cdk
cdk bootstrap
```

### Step 4: Deploy to AWS (10 minutes)

```bash
# Review what will be created
cdk diff

# Deploy!
cdk deploy

# Or from project root
cd ../..
make deploy-lambda
```

**Save the outputs:**
- API Gateway URL (e.g., `https://abc123.execute-api.us-east-1.amazonaws.com/prod`)
- S3 Bucket name
- Lambda function ARN

### Step 5: Configure Environment Variables (5 minutes)

#### Get your S3 bucket name from deployment outputs:
```bash
# From CDK output, you'll see something like:
# GimageStack.ImageBucketName = gimage-images-123456789012
```

#### Set environment variables in Lambda:
```bash
aws lambda update-function-configuration \
  --function-name GimageFunction \
  --environment "Variables={
    GEMINI_API_KEY=your-gemini-api-key-here,
    S3_BUCKET=gimage-images-123456789012,
    MAX_RESPONSE_SIZE_KB=512,
    LOG_LEVEL=info
  }"
```

**Or use AWS Console:**
1. Go to Lambda → Functions → GimageFunction
2. Configuration → Environment variables → Edit
3. Add:
   - `GEMINI_API_KEY` = your key
   - `S3_BUCKET` = bucket name from output
   - `MAX_RESPONSE_SIZE_KB` = 512
   - `LOG_LEVEL` = info

## ✅ Test Your Deployment (5 minutes)

### Test 1: Health Check

```bash
export API_URL="https://YOUR-API-ID.execute-api.REGION.amazonaws.com/prod"

curl $API_URL/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Test 2: Interactive Swagger UI

Open in browser:
```
https://YOUR-API-ID.execute-api.REGION.amazonaws.com/prod/docs
```

You should see beautiful interactive documentation!

### Test 3: Generate an Image

In Swagger UI:
1. Expand `POST /generate`
2. Click "Try it out"
3. Edit request body:
```json
{
  "prompt": "a sunset over mountains with vibrant colors",
  "size": "1024x1024",
  "response_format": "s3_url"
}
```
4. Click "Execute"
5. See your generated image URL in the response!

### Test 4: Resize an Image

```bash
# Get a test image as base64 (macOS)
BASE64_IMAGE=$(base64 -i some-image.jpg)

curl -X POST $API_URL/resize \
  -H "Content-Type: application/json" \
  -d "{
    \"image\": \"$BASE64_IMAGE\",
    \"width\": 800,
    \"height\": 600
  }"
```

## 🎉 You're Live!

Your Gimage Lambda API is now deployed and ready for production use!

### Next Steps

1. **Share with your team:**
   - API URL: `https://YOUR-API-ID.execute-api.REGION.amazonaws.com/prod`
   - Swagger UI: `https://YOUR-API-ID.execute-api.REGION.amazonaws.com/prod/docs`

2. **Integrate into your app:**
   - See [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) for TypeScript/Python/Go SDKs
   - Use the OpenAPI spec at `/openapi.yaml` to auto-generate clients

3. **Monitor your API:**
   ```bash
   # Tail logs
   make lambda-logs
   
   # Or
   aws logs tail /aws/lambda/GimageFunction --follow
   ```

4. **Set up custom domain** (optional):
   - See [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md#custom-domain-optional)

## 💰 Cost Estimate

For **10,000 requests/month**:
- Lambda Compute (ARM64): $0.17
- Lambda Requests: $0.002
- S3 Storage: $0.023
- S3 Requests: $0.01
- API Gateway: $0.035
- CloudWatch: $0.01
- **Total: ~$0.25/month**

Plus Gemini/Vertex AI costs (separate billing).

## 🐛 Troubleshooting

### Issue: Deployment fails with "bootstrap required"
**Solution:**
```bash
cd infrastructure/cdk
cdk bootstrap
```

### Issue: Lambda timeout
**Solution:** Increase timeout in `lib/gimage-stack.ts`:
```typescript
timeout: cdk.Duration.minutes(5)
```

### Issue: "Access Denied" on S3
**Solution:** Verify bucket name in environment variables matches CDK output.

### Issue: Gemini API returns 401
**Solution:** Verify API key is correct and has proper permissions.

## 📚 Full Documentation

For complete details, see:

- **[DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)** - Comprehensive deployment guide
- **[INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)** - Client SDK implementations
- **[API_REFERENCE.md](API_REFERENCE.md)** - Quick API reference
- **[lambda.md](lambda.md)** - Complete architecture documentation
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** - What was built

## 🎓 Example Use Cases

### Web App: Resize User Uploads
```typescript
const response = await fetch(`${API_URL}/resize`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    image: base64Image,
    width: 800,
    height: 600
  })
});
```

### E-commerce: Generate Product Images
```python
result = requests.post(f'{API_URL}/generate', json={
    'prompt': 'professional product photo of a watch',
    'size': '1024x1024',
    'style': 'photorealistic'
})
```

### Batch Processing: Create Thumbnails
```go
client := gimage.NewClient(apiURL, apiKey)
results, err := client.BatchProcess(gimage.BatchRequest{
    Operations: []gimage.Operation{
        {Type: "resize", Width: 200, Height: 200},
    },
    Images: imageList,
})
```

## 🚀 Ready for Production!

Your Gimage Lambda API is:
- ✅ Deployed on AWS
- ✅ Auto-scaling
- ✅ Cost-optimized (ARM64)
- ✅ Documented with Swagger UI
- ✅ Ready for integration

**Happy image processing!** 🎨
