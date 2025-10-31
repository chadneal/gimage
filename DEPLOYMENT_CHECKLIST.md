# Gimage Lambda Deployment Checklist

Complete step-by-step guide to deploy the Gimage Lambda API to AWS.

## ✅ Pre-Deployment Verification

### Build Artifacts
- [x] Lambda binary built (`bin/lambda/bootstrap` - 42MB)
- [x] Deployment package created (`bin/lambda.zip` - 17MB)
- [x] Architecture: linux/arm64
- [x] Runtime: provided.al2023

### Documentation
- [x] OpenAPI specification (`openapi.yaml`)
- [x] Integration guide (`INTEGRATION_GUIDE.md`)
- [x] API reference (`API_REFERENCE.md`)
- [x] Swagger setup guide (`SWAGGER_SETUP.md`)
- [x] Implementation plan (`lambda.md`)
- [x] Implementation summary (`IMPLEMENTATION_SUMMARY.md`)

### Code Completeness
- [x] All 8 operation handlers implemented
- [x] Swagger UI endpoint (`/docs`)
- [x] OpenAPI spec endpoint (`/openapi.yaml`)
- [x] Health check endpoint (`/health`)
- [x] S3 client with AWS SDK v2
- [x] CORS headers configured
- [x] Error handling implemented

---

## 🚀 Deployment Steps

### Step 1: Prerequisites (5 minutes)

#### Required Tools
```bash
# Check AWS CLI
aws --version
# Expected: aws-cli/2.x or higher

# Check Node.js
node --version
# Expected: v20.x or higher

# Check CDK
npm list -g aws-cdk
# If not installed: npm install -g aws-cdk
```

#### AWS Configuration
```bash
# Configure AWS credentials
aws configure
# Enter: Access Key ID, Secret Access Key, Region (e.g., us-east-1)

# Verify credentials
aws sts get-caller-identity
```

#### Environment Variables
Create `.env` file in project root:
```bash
# Required for AI image generation
GEMINI_API_KEY=your-gemini-api-key-here
# OR
VERTEX_PROJECT=your-gcp-project-id
VERTEX_LOCATION=us-central1
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# Optional configurations
MAX_RESPONSE_SIZE_KB=512
DEFAULT_IMAGE_SIZE=1024x1024
LOG_LEVEL=info
```

---

### Step 2: Create CDK Infrastructure (30 minutes)

#### Initialize CDK Project
```bash
# Create infrastructure directory
mkdir -p infrastructure/cdk
cd infrastructure/cdk

# Initialize CDK project
cdk init app --language typescript

# Install dependencies
npm install @aws-cdk/aws-lambda \
            @aws-cdk/aws-apigateway \
            @aws-cdk/aws-s3 \
            @aws-cdk/aws-iam \
            @aws-cdk/aws-logs
```

#### Create CDK Stack

Copy the CDK stack code from `lambda.md` (Section: "CDK Infrastructure") to:
`infrastructure/cdk/lib/gimage-stack.ts`

The stack includes:
- Lambda function (ARM64, provided.al2023)
- API Gateway REST API
- S3 bucket for image storage
- IAM roles and permissions
- CloudWatch log groups
- Environment variable configuration

#### Update CDK App

Edit `infrastructure/cdk/bin/cdk.ts`:
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

---

### Step 3: Build Lambda Package (5 minutes)

```bash
# Return to project root
cd ../..

# Clean previous builds
make clean-lambda

# Build Lambda binary for ARM64
make build-lambda

# Create deployment package
make package-lambda

# Verify package
ls -lh bin/lambda.zip
# Expected: ~17MB
```

---

### Step 4: Bootstrap CDK (First Time Only)

```bash
cd infrastructure/cdk

# Bootstrap CDK in your AWS account
cdk bootstrap aws://ACCOUNT-ID/REGION

# Example:
# cdk bootstrap aws://123456789012/us-east-1
```

---

### Step 5: Deploy to AWS (10 minutes)

```bash
# Review changes (dry run)
cdk diff

# Deploy stack
cdk deploy

# Or use Makefile from project root
cd ../..
make deploy-lambda
```

**During deployment, CDK will:**
1. Create S3 bucket for Lambda code
2. Upload `lambda.zip` to S3
3. Create Lambda function
4. Create API Gateway REST API
5. Configure routes and permissions
6. Create CloudWatch log groups
7. Output API URL

**Save the outputs:**
- API Gateway URL (e.g., `https://abc123.execute-api.us-east-1.amazonaws.com/prod`)
- S3 Bucket name
- Lambda function ARN

---

### Step 6: Configure Environment Variables (5 minutes)

#### Option A: Update via AWS Console
1. Go to AWS Lambda Console
2. Find function: `GimageFunction`
3. Configuration → Environment variables
4. Add required variables from your `.env` file

#### Option B: Update via AWS CLI
```bash
aws lambda update-function-configuration \
  --function-name GimageFunction \
  --environment "Variables={
    GEMINI_API_KEY=your-key,
    S3_BUCKET=gimage-images-ACCOUNT,
    MAX_RESPONSE_SIZE_KB=512
  }"
```

#### Option C: Update CDK Stack
Add environment variables in `gimage-stack.ts` and redeploy:
```typescript
environment: {
  GEMINI_API_KEY: process.env.GEMINI_API_KEY || '',
  S3_BUCKET: imageBucket.bucketName,
  // ... other variables
}
```

Then redeploy:
```bash
cd infrastructure/cdk
cdk deploy
```

---

### Step 7: Test Deployment (10 minutes)

#### Test Health Endpoint
```bash
export API_URL="https://YOUR-API-ID.execute-api.REGION.amazonaws.com/prod"

curl $API_URL/health
# Expected: {"status":"healthy","timestamp":"2025-10-31T..."}
```

#### Test Swagger UI
Open in browser:
```
https://YOUR-API-ID.execute-api.REGION.amazonaws.com/prod/docs
```

You should see interactive Swagger UI documentation.

#### Test Image Generation
```bash
curl -X POST $API_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "a sunset over mountains",
    "size": "1024x1024",
    "response_format": "s3_url"
  }'

# Expected: JSON with s3_url or base64 image
```

#### Test Image Resize
```bash
# First, get a test image as base64
BASE64_IMAGE=$(base64 -i test/fixtures/test-image.png)

curl -X POST $API_URL/resize \
  -H "Content-Type: application/json" \
  -d "{
    \"image\": \"$BASE64_IMAGE\",
    \"width\": 800,
    \"height\": 600
  }"
```

---

### Step 8: Monitor & Verify (5 minutes)

#### Check CloudWatch Logs
```bash
# Tail Lambda logs
make lambda-logs

# Or via AWS CLI
aws logs tail /aws/lambda/GimageFunction --follow
```

#### View Metrics
```bash
# Open CloudWatch dashboard
aws cloudwatch get-dashboard --dashboard-name GimageDashboard
```

#### Verify S3 Bucket
```bash
# List uploaded images
aws s3 ls s3://gimage-images-ACCOUNT/
```

---

## 🔧 Post-Deployment Configuration

### Custom Domain (Optional)

#### Using API Gateway Custom Domain
```bash
# Create certificate in ACM
aws acm request-certificate \
  --domain-name api.yourdomain.com \
  --validation-method DNS

# Create custom domain in API Gateway
aws apigateway create-domain-name \
  --domain-name api.yourdomain.com \
  --certificate-arn arn:aws:acm:...
```

#### Update DNS
Add CNAME record:
```
api.yourdomain.com → YOUR-API-ID.execute-api.REGION.amazonaws.com
```

---

### API Key Authentication (Optional)

#### Enable API Keys
Update CDK stack to add API key requirement:
```typescript
const apiKey = api.addApiKey('GimageApiKey', {
  apiKeyName: 'gimage-api-key',
  description: 'API key for Gimage API',
});

const usagePlan = api.addUsagePlan('GimageUsagePlan', {
  name: 'Standard',
  throttle: {
    rateLimit: 100,
    burstLimit: 200,
  },
  quota: {
    limit: 10000,
    period: apigateway.Period.MONTH,
  },
});

usagePlan.addApiKey(apiKey);
usagePlan.addApiStage({
  stage: api.deploymentStage,
});
```

Redeploy:
```bash
cdk deploy
```

Get API key:
```bash
aws apigateway get-api-keys --include-values
```

---

### CORS Configuration (Optional)

If you need to customize CORS settings, update `response.go`:

```go
func corsHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin": "https://yourdomain.com", // Specific domain
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type,X-Api-Key",
		"Access-Control-Max-Age": "86400",
	}
}
```

Rebuild and redeploy:
```bash
make build-lambda
make package-lambda
cdk deploy
```

---

### Monitoring & Alarms

#### Create CloudWatch Alarms
```bash
# High error rate alarm
aws cloudwatch put-metric-alarm \
  --alarm-name GimageHighErrorRate \
  --alarm-description "Alert when error rate > 5%" \
  --metric-name Errors \
  --namespace AWS/Lambda \
  --statistic Average \
  --period 300 \
  --threshold 0.05 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=FunctionName,Value=GimageFunction

# High duration alarm
aws cloudwatch put-metric-alarm \
  --alarm-name GimageHighDuration \
  --alarm-description "Alert when duration > 10 seconds" \
  --metric-name Duration \
  --namespace AWS/Lambda \
  --statistic Average \
  --period 300 \
  --threshold 10000 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=FunctionName,Value=GimageFunction
```

---

## 📊 Verification Checklist

### Functional Tests
- [ ] Health endpoint returns 200
- [ ] Swagger UI loads at `/docs`
- [ ] OpenAPI spec available at `/openapi.yaml`
- [ ] Image generation works (Gemini/Vertex)
- [ ] Image resize works
- [ ] Image scale works
- [ ] Image crop works
- [ ] Image compress works
- [ ] Image convert works
- [ ] Batch processing works
- [ ] CORS headers present
- [ ] Error responses formatted correctly

### Performance Tests
- [ ] Cold start < 3 seconds
- [ ] Warm execution < 500ms
- [ ] Large image processing < 5 seconds
- [ ] Batch operations complete within timeout
- [ ] S3 upload/download working
- [ ] Presigned URLs valid

### Security Tests
- [ ] S3 bucket not publicly accessible
- [ ] IAM roles follow least-privilege
- [ ] API keys work (if enabled)
- [ ] CORS configured correctly
- [ ] Environment variables encrypted
- [ ] CloudWatch logs sanitized (no secrets)

### Cost Verification
- [ ] Lambda using ARM64 (Graviton2)
- [ ] S3 lifecycle rules active (1-day deletion)
- [ ] CloudWatch logs retention set (7 days)
- [ ] No unnecessary Lambda invocations
- [ ] API Gateway caching configured (optional)

---

## 🐛 Troubleshooting

### Issue: Lambda Timeout
**Symptoms**: Requests timeout after 30 seconds

**Solution**: Increase timeout in CDK:
```typescript
timeout: cdk.Duration.minutes(5)
```

### Issue: Out of Memory
**Symptoms**: Lambda fails with OOM error

**Solution**: Increase memory in CDK:
```typescript
memorySize: 3008 // Increase from 2048
```

### Issue: Gemini API Key Not Working
**Symptoms**: 401 Unauthorized from Gemini

**Solution**:
1. Verify environment variable set correctly
2. Check API key validity at https://aistudio.google.com/app/apikey
3. Ensure key has permissions for Gemini 2.5 Flash Image

### Issue: S3 Access Denied
**Symptoms**: Cannot upload/download from S3

**Solution**:
1. Verify IAM role has S3 permissions
2. Check bucket policy
3. Ensure bucket exists in same region

### Issue: CORS Errors
**Symptoms**: Browser blocks requests

**Solution**:
1. Verify CORS headers in responses
2. Check `Access-Control-Allow-Origin`
3. Ensure OPTIONS method handled

### Issue: Swagger UI Not Loading
**Symptoms**: Blank page at `/docs`

**Solution**:
1. Check Lambda logs for errors
2. Verify route configured in API Gateway
3. Test OpenAPI spec endpoint first

---

## 📚 Next Steps After Deployment

### 1. Share API with Developers
- [ ] Distribute API URL
- [ ] Share Swagger UI link (`/docs`)
- [ ] Provide API keys (if enabled)
- [ ] Share integration guides

### 2. Set Up Monitoring
- [ ] Create CloudWatch dashboard
- [ ] Configure alarms
- [ ] Set up SNS notifications
- [ ] Enable X-Ray tracing (optional)

### 3. Documentation
- [ ] Update project README with API URL
- [ ] Add deployment info to wiki
- [ ] Document environment setup
- [ ] Create runbook for common issues

### 4. Client Integration
- [ ] Test TypeScript SDK
- [ ] Test Python SDK
- [ ] Test Go SDK
- [ ] Validate real-world use cases

### 5. Optimization
- [ ] Review CloudWatch metrics
- [ ] Optimize cold starts
- [ ] Adjust memory/timeout settings
- [ ] Enable caching if needed
- [ ] Consider provisioned concurrency for high traffic

---

## 💰 Cost Monitoring

### Set Up Budget Alerts
```bash
aws budgets create-budget \
  --account-id ACCOUNT-ID \
  --budget file://budget.json \
  --notifications-with-subscribers file://notifications.json
```

**budget.json:**
```json
{
  "BudgetName": "GimageMonthlyBudget",
  "BudgetLimit": {
    "Amount": "10",
    "Unit": "USD"
  },
  "TimeUnit": "MONTHLY",
  "BudgetType": "COST"
}
```

### Monitor Costs
```bash
# View current month costs
aws ce get-cost-and-usage \
  --time-period Start=2025-10-01,End=2025-10-31 \
  --granularity MONTHLY \
  --metrics UnblendedCost \
  --filter file://filter.json
```

---

## 📞 Support & Resources

### Documentation
- **OpenAPI Spec**: `openapi.yaml`
- **Integration Guide**: `INTEGRATION_GUIDE.md`
- **API Reference**: `API_REFERENCE.md`
- **Swagger Setup**: `SWAGGER_SETUP.md`
- **Lambda Plan**: `lambda.md`

### AWS Resources
- [Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [API Gateway Documentation](https://docs.aws.amazon.com/apigateway/)
- [CDK Documentation](https://docs.aws.amazon.com/cdk/)
- [AWS Support](https://aws.amazon.com/support/)

### Community
- **GitHub Issues**: https://github.com/chadneal/gimage/issues
- **Discussions**: https://github.com/chadneal/gimage/discussions

---

## ✅ Deployment Complete!

Once all checklist items are verified:

🎉 **Your Gimage Lambda API is live and ready for production use!**

**API Endpoints:**
- Health: `https://YOUR-API-URL/prod/health`
- Docs: `https://YOUR-API-URL/prod/docs`
- OpenAPI: `https://YOUR-API-URL/prod/openapi.yaml`

**Next:** Share the Swagger UI URL with your development team and start integrating!

---

**Deployment Date**: _____________
**Deployed By**: _____________
**API URL**: _____________
**Region**: _____________
**Notes**: _____________
