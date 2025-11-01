# Gimage AWS Lambda Distribution Plan

## Overview

This document outlines the comprehensive plan to convert the `gimage` CLI tool into a fully functional AWS Lambda-based API service. This will enable web applications like sophiestrophies.com to leverage remote compute for AI image generation and processing operations.

## Architecture

### High-Level Design

```
┌─────────────┐      ┌──────────────┐      ┌────────────────┐
│   Client    │─────▶│  API Gateway │─────▶│  Lambda (Go)   │
│  (Web App)  │◀─────│   REST API   │◀─────│    Handler     │
└─────────────┘      └──────────────┘      └────────────────┘
                                                    │
                                     ┌──────────────┼──────────────┐
                                     │              │              │
                                     ▼              ▼              ▼
                              ┌──────────┐   ┌──────────┐   ┌──────────┐
                              │    S3    │   │  Gemini  │   │  Vertex  │
                              │  Bucket  │   │   API    │   │   AI     │
                              └──────────┘   └──────────┘   └──────────┘
```

### Components

1. **Lambda Function** - Go 1.22+ runtime executing image operations
2. **API Gateway** - REST API with resource-based routing
3. **S3 Bucket** - Temporary storage for input/output images
4. **Gemini/Vertex APIs** - AI image generation services
5. **CloudWatch** - Logging and monitoring

### Request Flow

1. Client sends HTTP request to API Gateway with image operation details
2. API Gateway routes to Lambda handler
3. Lambda downloads source image (if needed) from S3 or base64 payload
4. Lambda executes image operation using existing gimage code
5. Lambda uploads result to S3 (for large images) or returns base64 (for small images)
6. Client receives response with image data or S3 presigned URL

## API Endpoints

### 1. Image Generation

**POST** `/generate`

Generate AI images from text prompts.

```json
{
  "prompt": "a sunset over mountains",
  "model": "gemini-2.5-flash-image",
  "size": "1024x1024",
  "style": "photorealistic",
  "negative_prompt": "people, buildings",
  "seed": 42,
  "response_format": "base64|s3_url"
}
```

Response:
```json
{
  "image": "base64_encoded_image_data",
  "s3_url": "https://s3.amazonaws.com/...",
  "width": 1024,
  "height": 1024,
  "format": "png",
  "size_bytes": 1048576
}
```

### 2. Image Resize

**POST** `/resize`

Resize images to specific dimensions.

```json
{
  "image": "base64_encoded_input|s3_key",
  "width": 800,
  "height": 600,
  "response_format": "base64|s3_url"
}
```

### 3. Image Scale

**POST** `/scale`

Scale images by a factor.

```json
{
  "image": "base64_encoded_input|s3_key",
  "factor": 0.5,
  "response_format": "base64|s3_url"
}
```

### 4. Image Crop

**POST** `/crop`

Crop images to specific region.

```json
{
  "image": "base64_encoded_input|s3_key",
  "x": 100,
  "y": 100,
  "width": 800,
  "height": 600,
  "response_format": "base64|s3_url"
}
```

### 5. Image Compress

**POST** `/compress`

Compress images to reduce file size.

```json
{
  "image": "base64_encoded_input|s3_key",
  "quality": 85,
  "format": "jpg|png|webp",
  "response_format": "base64|s3_url"
}
```

### 6. Image Convert

**POST** `/convert`

Convert images between formats.

```json
{
  "image": "base64_encoded_input|s3_key",
  "target_format": "webp",
  "response_format": "base64|s3_url"
}
```

### 7. Batch Processing

**POST** `/batch`

Process multiple images concurrently (async).

```json
{
  "operations": [
    {
      "operation": "resize",
      "image": "s3_key_1",
      "width": 800,
      "height": 600
    },
    {
      "operation": "compress",
      "image": "s3_key_2",
      "quality": 85
    }
  ],
  "callback_url": "https://myapp.com/webhook"
}
```

Response (async):
```json
{
  "batch_id": "batch-uuid-123",
  "status": "processing",
  "status_url": "/batch/batch-uuid-123"
}
```

### 8. Health Check

**GET** `/health`

Health check endpoint for monitoring.

```json
{
  "status": "healthy",
  "version": "0.1.1",
  "apis": {
    "gemini": "available",
    "vertex": "available"
  }
}
```

## Implementation Structure

### Directory Structure

```
gimage/
├── cmd/
│   ├── gimage/           # Existing CLI entrypoint
│   └── lambda/           # NEW: Lambda entrypoint
│       └── main.go       # Lambda handler main
├── internal/
│   ├── cli/              # Existing CLI commands
│   ├── imaging/          # Existing image processing (reuse)
│   ├── generate/         # Existing AI generation (reuse)
│   ├── config/           # Existing config (adapt for Lambda)
│   └── lambda/           # NEW: Lambda-specific code
│       ├── handler.go    # Main Lambda handler
│       ├── routes.go     # Route mapping
│       ├── dto.go        # Request/response DTOs
│       ├── s3.go         # S3 utilities
│       └── middleware.go # Logging, errors, CORS
├── infrastructure/       # NEW: CDK infrastructure
│   ├── cdk/
│   │   ├── lib/
│   │   │   └── gimage-stack.ts
│   │   ├── bin/
│   │   │   └── gimage.ts
│   │   ├── package.json
│   │   └── cdk.json
│   └── README.md
└── Makefile              # Updated with Lambda targets
```

### Key Implementation Files

#### 1. Lambda Handler (`cmd/lambda/main.go`)

```go
package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/apresai/gimage/internal/lambda"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler := lambda.NewHandler()
	return handler.Handle(ctx, req)
}
```

#### 2. Route Handler (`internal/lambda/handler.go`)

```go
package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/apresai/gimage/internal/imaging"
	"github.com/apresai/gimage/internal/generate"
)

type Handler struct {
	s3Client *S3Client
	routes   map[string]RouteHandler
}

type RouteHandler func(ctx context.Context, body []byte) (Response, error)

func NewHandler() *Handler {
	h := &Handler{
		s3Client: NewS3Client(),
		routes:   make(map[string]RouteHandler),
	}

	// Register routes
	h.registerRoutes()
	return h
}

func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Route mapping
	routeKey := fmt.Sprintf("%s %s", req.HTTPMethod, req.Path)

	handler, exists := h.routes[routeKey]
	if !exists {
		return errorResponse(404, "Route not found"), nil
	}

	// Execute handler
	response, err := handler(ctx, []byte(req.Body))
	if err != nil {
		return errorResponse(500, err.Error()), nil
	}

	// Return success
	return successResponse(response), nil
}

func (h *Handler) registerRoutes() {
	h.routes["POST /generate"] = h.handleGenerate
	h.routes["POST /resize"] = h.handleResize
	h.routes["POST /scale"] = h.handleScale
	h.routes["POST /crop"] = h.handleCrop
	h.routes["POST /compress"] = h.handleCompress
	h.routes["POST /convert"] = h.handleConvert
	h.routes["POST /batch"] = h.handleBatch
	h.routes["GET /health"] = h.handleHealth
}
```

#### 3. S3 Client (`internal/lambda/s3.go`)

```go
package lambda

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"os"
	"time"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client() *S3Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	return &S3Client{
		client: s3.NewFromConfig(cfg),
		bucket: os.Getenv("S3_BUCKET"),
	}
}

func (s *S3Client) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	return err
}

func (s *S3Client) Download(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	return buf.Bytes(), err
}

func (s *S3Client) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}
```

#### 4. DTOs (`internal/lambda/dto.go`)

```go
package lambda

type GenerateRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	Size           string `json:"size"`
	Style          string `json:"style"`
	NegativePrompt string `json:"negative_prompt"`
	Seed           int64  `json:"seed"`
	ResponseFormat string `json:"response_format"` // "base64" or "s3_url"
}

type ResizeRequest struct {
	Image          string `json:"image"` // base64 or s3_key
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	ResponseFormat string `json:"response_format"`
}

type ImageResponse struct {
	Image     string `json:"image,omitempty"`      // base64 encoded
	S3URL     string `json:"s3_url,omitempty"`     // presigned URL
	S3Key     string `json:"s3_key,omitempty"`     // S3 key for chaining
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	SizeBytes int64  `json:"size_bytes"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
```

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `S3_BUCKET` | S3 bucket name for temporary image storage | `gimage-storage-prod` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `GEMINI_API_KEY` | Google Gemini API key for AI generation | `AIza...` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `VERTEX_API_KEY` | Vertex AI API key (Express Mode) | - |
| `VERTEX_PROJECT` | GCP project ID for Vertex AI | - |
| `VERTEX_LOCATION` | Vertex AI location | `us-central1` |
| `GOOGLE_APPLICATION_CREDENTIALS_BASE64` | Base64-encoded service account JSON for Vertex AI Full Mode | - |
| `MAX_IMAGE_SIZE_MB` | Maximum image size to process | `10` |
| `MAX_RESPONSE_SIZE_KB` | Max size for base64 response (larger → S3) | `512` |
| `PRESIGNED_URL_EXPIRATION_MINUTES` | S3 presigned URL expiration | `60` |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `ENABLE_CORS` | Enable CORS headers | `true` |
| `ALLOWED_ORIGINS` | Comma-separated allowed origins | `*` |

### Secret Management

For production deployments, use AWS Secrets Manager or SSM Parameter Store for sensitive values:

```typescript
// In CDK stack
const geminiApiKey = secretsmanager.Secret.fromSecretNameV2(this, 'GeminiKey', 'gimage/gemini-api-key');

lambdaFunction.addEnvironment('GEMINI_API_KEY', geminiApiKey.secretValue.toString());
```

## CDK Infrastructure

### Stack Definition (`infrastructure/cdk/lib/gimage-stack.ts`)

```typescript
import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';

export class GimageStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // S3 Bucket for temporary image storage
    const imageBucket = new s3.Bucket(this, 'GimageStorage', {
      bucketName: `gimage-storage-${this.account}-${this.region}`,
      encryption: s3.BucketEncryption.S3_MANAGED,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      lifecycleRules: [
        {
          id: 'DeleteOldImages',
          enabled: true,
          expiration: cdk.Duration.days(1), // Auto-delete after 1 day
        },
      ],
      cors: [
        {
          allowedMethods: [s3.HttpMethods.GET, s3.HttpMethods.PUT],
          allowedOrigins: ['*'], // Configure per environment
          allowedHeaders: ['*'],
          maxAge: 3000,
        },
      ],
    });

    // Lambda Function
    const gimageFunction = new lambda.Function(this, 'GimageFunction', {
      functionName: 'gimage-processor',
      runtime: lambda.Runtime.PROVIDED_AL2023, // Custom Go runtime
      handler: 'bootstrap', // Go Lambda requires "bootstrap" as handler name
      code: lambda.Code.fromAsset('../../bin/lambda.zip'), // Packaged Lambda
      architecture: lambda.Architecture.ARM_64, // Graviton2 for cost savings
      memorySize: 2048, // 2GB for image processing
      timeout: cdk.Duration.minutes(5), // Max for API Gateway
      environment: {
        S3_BUCKET: imageBucket.bucketName,
        AWS_REGION: this.region,
        LOG_LEVEL: 'info',
        MAX_IMAGE_SIZE_MB: '10',
        MAX_RESPONSE_SIZE_KB: '512',
        PRESIGNED_URL_EXPIRATION_MINUTES: '60',
        ENABLE_CORS: 'true',
        ALLOWED_ORIGINS: '*',
      },
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    // Grant S3 permissions to Lambda
    imageBucket.grantReadWrite(gimageFunction);

    // API Gateway
    const api = new apigateway.RestApi(this, 'GimageApi', {
      restApiName: 'Gimage API',
      description: 'AI-powered image generation and processing API',
      deployOptions: {
        stageName: 'prod',
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
        metricsEnabled: true,
      },
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ['Content-Type', 'Authorization'],
      },
    });

    // Lambda integration
    const lambdaIntegration = new apigateway.LambdaIntegration(gimageFunction, {
      proxy: true,
    });

    // API Routes
    const generateResource = api.root.addResource('generate');
    generateResource.addMethod('POST', lambdaIntegration);

    const resizeResource = api.root.addResource('resize');
    resizeResource.addMethod('POST', lambdaIntegration);

    const scaleResource = api.root.addResource('scale');
    scaleResource.addMethod('POST', lambdaIntegration);

    const cropResource = api.root.addResource('crop');
    cropResource.addMethod('POST', lambdaIntegration);

    const compressResource = api.root.addResource('compress');
    compressResource.addMethod('POST', lambdaIntegration);

    const convertResource = api.root.addResource('convert');
    convertResource.addMethod('POST', lambdaIntegration);

    const batchResource = api.root.addResource('batch');
    batchResource.addMethod('POST', lambdaIntegration);

    const healthResource = api.root.addResource('health');
    healthResource.addMethod('GET', lambdaIntegration);

    // Documentation endpoints
    const docsResource = api.root.addResource('docs');
    docsResource.addMethod('GET', lambdaIntegration);

    const openapiResource = api.root.addResource('openapi.yaml');
    openapiResource.addMethod('GET', lambdaIntegration);

    // Outputs
    new cdk.CfnOutput(this, 'ApiUrl', {
      value: api.url,
      description: 'Gimage API Gateway URL',
    });

    new cdk.CfnOutput(this, 'BucketName', {
      value: imageBucket.bucketName,
      description: 'S3 bucket for image storage',
    });

    new cdk.CfnOutput(this, 'FunctionArn', {
      value: gimageFunction.functionArn,
      description: 'Lambda function ARN',
    });
  }
}
```

### App Entry Point (`infrastructure/cdk/bin/gimage.ts`)

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
  description: 'Gimage - AI-powered image generation and processing API',
});

app.synth();
```

### CDK Configuration (`infrastructure/cdk/cdk.json`)

```json
{
  "app": "npx ts-node --prefer-ts-exts bin/gimage.ts",
  "watch": {
    "include": ["**"],
    "exclude": [
      "README.md",
      "cdk*.json",
      "**/*.d.ts",
      "**/*.js",
      "tsconfig.json",
      "package*.json",
      "yarn.lock",
      "node_modules",
      "test"
    ]
  },
  "context": {
    "@aws-cdk/aws-lambda:recognizeLayerVersion": true,
    "@aws-cdk/core:checkSecretUsage": true,
    "@aws-cdk/core:target-partitions": ["aws", "aws-cn"],
    "@aws-cdk-containers/ecs-service-extensions:enableDefaultLogDriver": true,
    "@aws-cdk/aws-ec2:uniqueImdsv2TemplateName": true,
    "@aws-cdk/aws-ecs:arnFormatIncludesClusterName": true,
    "@aws-cdk/aws-iam:minimizePolicies": true,
    "@aws-cdk/core:validateSnapshotRemovalPolicy": true,
    "@aws-cdk/aws-codepipeline:crossAccountKeyAliasStackSafeResourceName": true,
    "@aws-cdk/aws-s3:createDefaultLoggingPolicy": true,
    "@aws-cdk/aws-sns-subscriptions:restrictSqsDescryption": true,
    "@aws-cdk/aws-apigateway:disableCloudWatchRole": false,
    "@aws-cdk/core:enablePartitionLiterals": true,
    "@aws-cdk/aws-events:eventsTargetQueueSameAccount": true,
    "@aws-cdk/aws-iam:standardizedServicePrincipals": true,
    "@aws-cdk/aws-ecs:disableExplicitDeploymentControllerForCircuitBreaker": true,
    "@aws-cdk/aws-iam:importedRoleStackSafeDefaultPolicyName": true,
    "@aws-cdk/aws-s3:serverAccessLogsUseBucketPolicy": true,
    "@aws-cdk/aws-route53-patternslambda:useDotDelimitedDomainName": true,
    "@aws-cdk/customresources:installLatestAwsSdkDefault": false,
    "@aws-cdk/aws-rds:databaseProxyUniqueResourceName": true,
    "@aws-cdk/aws-codedeploy:removeAlarmsFromDeploymentGroup": true,
    "@aws-cdk/aws-apigateway:authorizerChangeDeploymentLogicalId": true,
    "@aws-cdk/aws-ec2:launchTemplateDefaultUserData": true,
    "@aws-cdk/aws-secretsmanager:useAttachedSecretResourcePolicyForSecretTargetAttachments": true,
    "@aws-cdk/aws-redshift:columnId": true,
    "@aws-cdk/aws-stepfunctions-tasks:enableEmrServicePolicyV2": true,
    "@aws-cdk/aws-ec2:restrictDefaultSecurityGroup": true,
    "@aws-cdk/aws-apigateway:requestValidatorUniqueId": true,
    "@aws-cdk/aws-kms:aliasNameRef": true,
    "@aws-cdk/aws-autoscaling:generateLaunchTemplateInsteadOfLaunchConfig": true,
    "@aws-cdk/core:includePrefixInUniqueNameGeneration": true,
    "@aws-cdk/aws-opensearchservice:enableOpensearchMultiAzWithStandby": true
  }
}
```

### Package Configuration (`infrastructure/cdk/package.json`)

```json
{
  "name": "gimage-cdk",
  "version": "0.1.1",
  "description": "CDK infrastructure for Gimage Lambda API",
  "bin": {
    "gimage": "bin/gimage.js"
  },
  "scripts": {
    "build": "tsc",
    "watch": "tsc -w",
    "test": "jest",
    "cdk": "cdk",
    "synth": "cdk synth",
    "deploy": "cdk deploy",
    "diff": "cdk diff",
    "destroy": "cdk destroy"
  },
  "devDependencies": {
    "@types/jest": "^29.5.0",
    "@types/node": "20.10.0",
    "jest": "^29.5.0",
    "ts-jest": "^29.1.0",
    "aws-cdk": "2.120.0",
    "ts-node": "^10.9.1",
    "typescript": "~5.3.0"
  },
  "dependencies": {
    "aws-cdk-lib": "2.120.0",
    "constructs": "^10.0.0",
    "source-map-support": "^0.5.21"
  }
}
```

## Build System

### Updated Makefile

Add Lambda-specific targets to the existing Makefile:

```makefile
# Lambda-specific targets
.PHONY: build-lambda package-lambda deploy-lambda clean-lambda

## build-lambda: Build Lambda function binary for AWS
build-lambda:
	@echo "Building Lambda function for AWS ARM64..."
	@mkdir -p bin/lambda
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) \
		-tags lambda.norpc \
		-o bin/lambda/bootstrap \
		./cmd/lambda
	@chmod +x bin/lambda/bootstrap
	@echo "Lambda binary built: bin/lambda/bootstrap"

## package-lambda: Package Lambda function for deployment
package-lambda: build-lambda
	@echo "Packaging Lambda function..."
	@cd bin/lambda && zip -r ../lambda.zip bootstrap
	@echo "Lambda package created: bin/lambda.zip ($(shell du -h bin/lambda.zip | cut -f1))"

## deploy-lambda: Deploy Lambda function using CDK
deploy-lambda: package-lambda
	@echo "Deploying Lambda function with CDK..."
	@cd infrastructure/cdk && npm run build && npm run deploy

## clean-lambda: Clean Lambda build artifacts
clean-lambda:
	@echo "Cleaning Lambda artifacts..."
	@rm -rf bin/lambda bin/lambda.zip
	@echo "Lambda artifacts cleaned"

## lambda-logs: Tail Lambda function logs
lambda-logs:
	@echo "Tailing Lambda logs..."
	@aws logs tail /aws/lambda/gimage-processor --follow

## lambda-invoke: Invoke Lambda function locally for testing
lambda-invoke: build-lambda
	@echo "Invoking Lambda locally..."
	@cd bin/lambda && \
		GEMINI_API_KEY=$(GEMINI_API_KEY) \
		S3_BUCKET=test-bucket \
		./bootstrap
```

### GitHub Actions Workflow (`.github/workflows/lambda-deploy.yml`)

```yaml
name: Deploy Lambda

on:
  push:
    branches: [main]
    paths:
      - 'cmd/lambda/**'
      - 'internal/lambda/**'
      - 'infrastructure/cdk/**'
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - name: Build Lambda binary
        run: make build-lambda

      - name: Package Lambda
        run: make package-lambda

      - name: Install CDK dependencies
        working-directory: infrastructure/cdk
        run: npm ci

      - name: Deploy with CDK
        working-directory: infrastructure/cdk
        run: |
          npm run build
          npm run deploy -- --require-approval never
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
          VERTEX_API_KEY: ${{ secrets.VERTEX_API_KEY }}
          VERTEX_PROJECT: ${{ secrets.VERTEX_PROJECT }}
```

## Testing Strategy

### 1. Unit Tests

Test Lambda handlers independently:

```go
// internal/lambda/handler_test.go
package lambda

import (
	"context"
	"encoding/json"
	"testing"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestHandleResize(t *testing.T) {
	handler := NewHandler()

	request := ResizeRequest{
		Image:  "base64_encoded_test_image",
		Width:  800,
		Height: 600,
	}

	body, _ := json.Marshal(request)
	response, err := handler.handleResize(context.Background(), body)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Image)
	assert.Equal(t, 800, response.Width)
	assert.Equal(t, 600, response.Height)
}
```

### 2. Integration Tests

Test with actual S3 and APIs:

```bash
# Set environment variables
export S3_BUCKET=gimage-test-bucket
export GEMINI_API_KEY=your_test_key

# Run integration tests
go test -tags=integration ./internal/lambda/...
```

### 3. Local Testing with SAM CLI

```bash
# Install AWS SAM CLI
brew install aws-sam-cli

# Create SAM template (sam-template.yaml)
# Test locally
sam local invoke GimageFunction --event test-events/resize.json
```

### 4. Load Testing

Use Artillery or k6 for load testing:

```yaml
# load-test.yml (Artillery)
config:
  target: "https://api-id.execute-api.us-east-1.amazonaws.com/prod"
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 120
      arrivalRate: 50
      name: "Load test"

scenarios:
  - name: "Resize image"
    flow:
      - post:
          url: "/resize"
          json:
            image: "{{base64_test_image}}"
            width: 800
            height: 600
```

## Performance Optimization

### 1. Cold Start Mitigation

- Use ARM64 architecture (Graviton2) for faster execution
- Enable provisioned concurrency for critical paths
- Minimize binary size with build tags: `-tags lambda.norpc`
- Use Lambda SnapStart (when available for Go)

### 2. Memory and Timeout

- Recommended memory: 2048 MB (balances CPU and cost)
- Max timeout: 5 minutes (API Gateway limit)
- For longer operations, use async processing with Step Functions

### 3. Concurrency

- Set reserved concurrency per environment
- Production: 100 concurrent executions
- Development: 10 concurrent executions

### 4. Caching

- Cache frequently used images in S3 with predictable keys
- Use CloudFront CDN for serving processed images
- Implement in-memory caching for API responses

## Cost Estimation

### Lambda Pricing (us-east-1)

- **Compute**: $0.0000133334 per GB-second
- **Requests**: $0.20 per 1M requests

### Example Monthly Cost (10,000 requests)

| Operation | Duration | Memory | Cost |
|-----------|----------|--------|------|
| Generate (Gemini) | 3s | 2048 MB | $0.10 |
| Resize | 0.5s | 2048 MB | $0.02 |
| Compress | 0.8s | 2048 MB | $0.03 |
| Convert | 0.4s | 2048 MB | $0.02 |
| **Total Lambda** | - | - | **$0.17** |
| **Requests (10K)** | - | - | **$0.002** |
| **S3 Storage (1GB)** | - | - | **$0.023** |
| **S3 Requests** | - | - | **$0.01** |
| **Total Monthly** | - | - | **~$0.20** |

**Note**: Gemini/Vertex AI costs are separate and vary by model.

## Security Considerations

### 1. API Authentication

Add API key authentication:

```typescript
// In CDK stack
const apiKey = api.addApiKey('GimageApiKey', {
  apiKeyName: 'gimage-api-key',
});

const plan = api.addUsagePlan('GimageUsagePlan', {
  name: 'Standard',
  throttle: {
    rateLimit: 100,
    burstLimit: 200,
  },
  quota: {
    limit: 10000,
    period: apigateway.Period.DAY,
  },
});

plan.addApiKey(apiKey);
plan.addApiStage({ stage: api.deploymentStage });
```

### 2. IAM Permissions

Principle of least privilege:

```typescript
gimageFunction.addToRolePolicy(new iam.PolicyStatement({
  effect: iam.Effect.ALLOW,
  actions: [
    's3:GetObject',
    's3:PutObject',
  ],
  resources: [
    `${imageBucket.bucketArn}/*`,
  ],
}));
```

### 3. Input Validation

- Validate all request inputs
- Limit image sizes (max 10 MB by default)
- Sanitize file names and paths
- Rate limiting per API key

### 4. Secrets Management

```bash
# Store secrets in AWS Secrets Manager
aws secretsmanager create-secret \
  --name gimage/gemini-api-key \
  --secret-string "AIza..."

# Reference in Lambda
const secret = secretsmanager.Secret.fromSecretNameV2(
  this, 'GeminiKey', 'gimage/gemini-api-key'
);
```

## Monitoring and Observability

### CloudWatch Dashboards

Create custom dashboard:

```typescript
const dashboard = new cloudwatch.Dashboard(this, 'GimageDashboard', {
  dashboardName: 'gimage-metrics',
});

dashboard.addWidgets(
  new cloudwatch.GraphWidget({
    title: 'Lambda Invocations',
    left: [gimageFunction.metricInvocations()],
  }),
  new cloudwatch.GraphWidget({
    title: 'Lambda Errors',
    left: [gimageFunction.metricErrors()],
  }),
  new cloudwatch.GraphWidget({
    title: 'Lambda Duration',
    left: [gimageFunction.metricDuration()],
  }),
);
```

### CloudWatch Alarms

```typescript
gimageFunction.metricErrors().createAlarm(this, 'ErrorAlarm', {
  threshold: 10,
  evaluationPeriods: 2,
  alarmDescription: 'Alert when error rate is high',
});
```

### X-Ray Tracing

Enable in CDK:

```typescript
const gimageFunction = new lambda.Function(this, 'GimageFunction', {
  // ... other props
  tracing: lambda.Tracing.ACTIVE,
});
```

## Client SDK Examples

### JavaScript/TypeScript

```typescript
import axios from 'axios';

const API_URL = 'https://your-api-id.execute-api.us-east-1.amazonaws.com/prod';
const API_KEY = 'your-api-key';

export class GimageClient {
  async generateImage(prompt: string, options?: GenerateOptions): Promise<ImageResponse> {
    const response = await axios.post(`${API_URL}/generate`, {
      prompt,
      ...options,
      response_format: 's3_url', // Use S3 for large images
    }, {
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
    });

    return response.data;
  }

  async resizeImage(image: string, width: number, height: number): Promise<ImageResponse> {
    const response = await axios.post(`${API_URL}/resize`, {
      image, // Can be base64 or S3 key
      width,
      height,
      response_format: 'base64', // Small images as base64
    }, {
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
    });

    return response.data;
  }

  async compressImage(image: string, quality: number = 85): Promise<ImageResponse> {
    const response = await axios.post(`${API_URL}/compress`, {
      image,
      quality,
      format: 'webp', // Convert to WebP for web
      response_format: 's3_url',
    }, {
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
    });

    return response.data;
  }
}
```

### React Hook

```typescript
import { useState } from 'react';
import { GimageClient } from './gimage-client';

export function useGimage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const client = new GimageClient();

  const generateImage = async (prompt: string) => {
    setLoading(true);
    setError(null);

    try {
      const result = await client.generateImage(prompt);
      return result;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const resizeImage = async (file: File, width: number, height: number) => {
    setLoading(true);
    setError(null);

    try {
      // Convert file to base64
      const base64 = await fileToBase64(file);
      const result = await client.resizeImage(base64, width, height);
      return result;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { generateImage, resizeImage, loading, error };
}

async function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}
```

## Deployment Instructions

### Prerequisites

1. AWS CLI configured with appropriate credentials
2. Node.js 20+ and npm
3. Go 1.22+
4. AWS CDK CLI: `npm install -g aws-cdk`

### Initial Setup

```bash
# 1. Install CDK dependencies
cd infrastructure/cdk
npm install

# 2. Bootstrap CDK (first time only)
cdk bootstrap aws://ACCOUNT-ID/REGION

# 3. Set environment variables
export GEMINI_API_KEY=your_gemini_key
export VERTEX_API_KEY=your_vertex_key  # Optional
export VERTEX_PROJECT=your_gcp_project  # Optional
```

### Build and Deploy

```bash
# From project root

# 1. Build Lambda binary
make build-lambda

# 2. Package Lambda
make package-lambda

# 3. Preview changes
cd infrastructure/cdk
npm run diff

# 4. Deploy
npm run deploy

# 5. Get API endpoint
aws cloudformation describe-stacks \
  --stack-name GimageStack \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiUrl`].OutputValue' \
  --output text
```

### Update Deployment

```bash
# Rebuild and redeploy
make package-lambda
cd infrastructure/cdk
npm run deploy
```

### Rollback

```bash
# Destroy stack
cd infrastructure/cdk
npm run destroy
```

## Migration Path

### Phase 1: Core Lambda Infrastructure (Week 1)
- [ ] Create Lambda handler structure
- [ ] Implement S3 client
- [ ] Build route mapping
- [ ] Create DTOs for requests/responses
- [ ] Write CDK infrastructure

### Phase 2: Basic Operations (Week 2)
- [ ] Implement `/resize` endpoint
- [ ] Implement `/scale` endpoint
- [ ] Implement `/crop` endpoint
- [ ] Implement `/compress` endpoint
- [ ] Implement `/convert` endpoint

### Phase 3: AI Generation (Week 3)
- [ ] Implement `/generate` endpoint (Gemini)
- [ ] Add Vertex AI support
- [ ] Handle API authentication
- [ ] Add model selection

### Phase 4: Advanced Features (Week 4)
- [ ] Implement `/batch` endpoint with async processing
- [ ] Add health check endpoint
- [ ] Implement presigned URL generation
- [ ] Add response format negotiation (base64 vs S3)

### Phase 5: Production Hardening (Week 5)
- [ ] Add comprehensive error handling
- [ ] Implement request validation
- [ ] Add rate limiting
- [ ] Configure API authentication
- [ ] Setup CloudWatch dashboards and alarms

### Phase 6: Testing & Documentation (Week 6)
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Create client SDK examples
- [ ] Write API documentation
- [ ] Create deployment runbooks

## Future Enhancements

### 1. Async Processing with Step Functions

For operations taking >5 minutes:

```typescript
const stateMachine = new stepfunctions.StateMachine(this, 'GimageWorkflow', {
  definition: stepfunctions.Chain
    .start(new tasks.LambdaInvoke(this, 'ProcessImage', {
      lambdaFunction: gimageFunction,
    }))
    .next(new tasks.LambdaInvoke(this, 'NotifyCompletion', {
      lambdaFunction: notificationFunction,
    })),
});
```

### 2. GraphQL API

Use AppSync for more flexible queries:

```graphql
type Query {
  health: HealthStatus!
}

type Mutation {
  generateImage(input: GenerateInput!): ImageResult!
  resizeImage(input: ResizeInput!): ImageResult!
  batchProcess(input: BatchInput!): BatchResult!
}

type Subscription {
  onBatchComplete(batchId: ID!): BatchResult!
}
```

### 3. WebSocket Support

Real-time progress updates for long operations.

### 4. CDN Integration

```typescript
const distribution = new cloudfront.Distribution(this, 'ImageCDN', {
  defaultBehavior: {
    origin: new origins.S3Origin(imageBucket),
  },
});
```

### 5. Multi-Region Deployment

Deploy to multiple regions for global low-latency access.

## Conclusion

This plan provides a comprehensive roadmap for converting the gimage CLI tool into a production-ready AWS Lambda-based API service. The architecture maintains the purity of the existing Go codebase while adding the flexibility and scalability of serverless computing.

Key benefits:
- **Zero infrastructure management** - AWS handles scaling
- **Pay-per-use pricing** - Cost-effective for variable workloads
- **Global availability** - Deploy to multiple regions
- **Easy integration** - RESTful API with JSON
- **Type safety** - Existing Go code with strong typing
- **Production-ready** - Monitoring, logging, error handling

The implementation preserves all existing image processing and AI generation capabilities while making them accessible to web applications through a simple HTTP API.
