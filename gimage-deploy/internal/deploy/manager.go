package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/models"
	"github.com/apresai/gimage-deploy/internal/storage"
	awsConfig "github.com/aws/aws-sdk-go-v2/aws"
)

// Manager orchestrates deployment operations
type Manager struct {
	cfg            awsConfig.Config
	lambdaClient   *aws.LambdaClient
	s3Client       *aws.S3Client
	iamClient      *aws.IAMClient
	agClient       *aws.APIGatewayClient
	deploymentMgr  *storage.DeploymentManager
}

// NewManager creates a new deployment manager
func NewManager(cfg awsConfig.Config) *Manager {
	return &Manager{
		cfg:           cfg,
		lambdaClient:  aws.NewLambdaClient(cfg),
		s3Client:      aws.NewS3Client(cfg),
		iamClient:     aws.NewIAMClient(cfg),
		agClient:      aws.NewAPIGatewayClient(cfg),
		deploymentMgr: storage.NewDeploymentManager(),
	}
}

// DeployInput contains parameters for creating a deployment
type DeployInput struct {
	ID             string
	Stage          string
	Region         string
	MemoryMB       int
	TimeoutSec     int
	Concurrency    int
	Architecture   string
	Environment    map[string]string
	Description    string
	LambdaCodePath string // Path to Lambda deployment package (zip)
}

// Deploy creates a new deployment
func (m *Manager) Deploy(ctx context.Context, input DeployInput) (*models.Deployment, error) {
	// Load existing deployments
	if err := m.deploymentMgr.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deployments: %w", err)
	}

	// Check if deployment already exists
	if m.deploymentMgr.Exists(input.ID) {
		return nil, fmt.Errorf("deployment with ID %s already exists", input.ID)
	}

	region := aws.GetRegion(m.cfg)

	// Generate resource names
	bucketName := fmt.Sprintf("gimage-storage-%s", input.ID)
	functionName := fmt.Sprintf("gimage-processor-%s", input.ID)
	roleName := fmt.Sprintf("gimage-lambda-role-%s", input.ID)
	apiName := fmt.Sprintf("gimage-api-%s", input.ID)

	fmt.Printf("Creating deployment %s...\n", input.ID)

	// Step 1: Create S3 bucket
	fmt.Printf("  [1/6] Creating S3 bucket: %s\n", bucketName)
	if err := m.createS3Bucket(ctx, bucketName, region); err != nil {
		return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Step 2: Create IAM role
	fmt.Printf("  [2/6] Creating IAM role: %s\n", roleName)
	roleArn, err := m.createIAMRole(ctx, roleName, bucketName)
	if err != nil {
		// Cleanup S3 bucket
		m.s3Client.EmptyBucket(ctx, bucketName)
		m.s3Client.DeleteBucket(ctx, bucketName)
		return nil, fmt.Errorf("failed to create IAM role: %w", err)
	}

	// Wait for IAM role to propagate
	fmt.Printf("  [3/6] Waiting for IAM role to propagate (10s)...\n")
	time.Sleep(10 * time.Second)

	// Step 3: Create Lambda function
	fmt.Printf("  [4/6] Creating Lambda function: %s\n", functionName)
	lambdaArn, err := m.createLambdaFunction(ctx, functionName, roleArn, input)
	if err != nil {
		// Cleanup
		m.iamClient.DeleteRole(ctx, roleName)
		m.s3Client.EmptyBucket(ctx, bucketName)
		m.s3Client.DeleteBucket(ctx, bucketName)
		return nil, fmt.Errorf("failed to create Lambda function: %w", err)
	}

	// Step 4: Create API Gateway
	fmt.Printf("  [5/6] Creating API Gateway: %s\n", apiName)
	apiGatewayID, apiURL, err := m.createAPIGateway(ctx, apiName, lambdaArn, functionName, input.Stage, region)
	if err != nil {
		// Cleanup
		m.lambdaClient.DeleteFunction(ctx, functionName)
		m.iamClient.DeleteRole(ctx, roleName)
		m.s3Client.EmptyBucket(ctx, bucketName)
		m.s3Client.DeleteBucket(ctx, bucketName)
		return nil, fmt.Errorf("failed to create API Gateway: %w", err)
	}

	// Step 5: Save deployment to registry
	fmt.Printf("  [6/6] Saving deployment configuration...\n")
	deployment := &models.Deployment{
		ID:            input.ID,
		Stage:         input.Stage,
		Region:        region,
		FunctionName:  functionName,
		FunctionARN:   lambdaArn,
		APIGatewayID:  apiGatewayID,
		APIGatewayURL: apiURL,
		S3Bucket:      bucketName,
		IAMRoleARN:    roleArn,
		Version:       "1.2.63", // TODO: Get from actual gimage version
		Status:        models.StatusActive,
		Health: models.HealthStatus{
			IsHealthy:   true,
			Score:       100,
			LastChecked: time.Now(),
		},
		Configuration: models.LambdaConfiguration{
			MemoryMB:       input.MemoryMB,
			TimeoutSeconds: input.TimeoutSec,
			Concurrency:    input.Concurrency,
			Architecture:   input.Architecture,
			Runtime:        "provided.al2023",
			Handler:        "bootstrap",
		},
		EnvironmentVars: input.Environment,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := m.deploymentMgr.Add(deployment); err != nil {
		return nil, fmt.Errorf("failed to save deployment: %w", err)
	}

	fmt.Printf("\n✓ Deployment %s created successfully!\n", input.ID)
	fmt.Printf("  Endpoint: %s\n", apiURL)

	return deployment, nil
}

// createS3Bucket creates and configures an S3 bucket
func (m *Manager) createS3Bucket(ctx context.Context, bucketName, region string) error {
	// Check if bucket exists
	exists, err := m.s3Client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("bucket %s already exists", bucketName)
	}

	// Create bucket
	if err := m.s3Client.CreateBucket(ctx, bucketName, region); err != nil {
		return err
	}

	// Configure CORS
	if err := m.s3Client.PutBucketCORS(ctx, bucketName); err != nil {
		return err
	}

	// Configure lifecycle policy (delete images after 30 days)
	if err := m.s3Client.PutBucketLifecycle(ctx, bucketName, 30); err != nil {
		return err
	}

	// Block public access
	if err := m.s3Client.BlockPublicAccess(ctx, bucketName); err != nil {
		return err
	}

	return nil
}

// createIAMRole creates an IAM role with required policies
func (m *Manager) createIAMRole(ctx context.Context, roleName, bucketName string) (string, error) {
	// Create role
	roleArn, err := m.iamClient.CreateLambdaExecutionRole(ctx, roleName)
	if err != nil {
		return "", err
	}

	// Attach basic execution policy (CloudWatch Logs)
	if err := m.iamClient.AttachLambdaBasicExecutionPolicy(ctx, roleName); err != nil {
		return "", err
	}

	// Attach S3 access policy
	if err := m.iamClient.CreateS3AccessPolicy(ctx, roleName, bucketName); err != nil {
		return "", err
	}

	// Attach Bedrock access policy (optional, for AI generation)
	if err := m.iamClient.CreateBedrockAccessPolicy(ctx, roleName); err != nil {
		return "", err
	}

	return roleArn, nil
}

// createLambdaFunction creates a Lambda function
func (m *Manager) createLambdaFunction(ctx context.Context, functionName, roleArn string, input DeployInput) (string, error) {
	// Read Lambda code from file
	var codeBytes []byte
	var err error

	if input.LambdaCodePath != "" {
		codeBytes, err = os.ReadFile(input.LambdaCodePath)
		if err != nil {
			return "", fmt.Errorf("failed to read Lambda code: %w", err)
		}
	} else {
		// Use bootstrap from current directory or parent
		bootstrapPath := findBootstrapFile()
		if bootstrapPath == "" {
			return "", fmt.Errorf("Lambda code not found (specify --lambda-code or ensure bootstrap file exists)")
		}
		codeBytes, err = createDeploymentZip(bootstrapPath)
		if err != nil {
			return "", fmt.Errorf("failed to create deployment zip: %w", err)
		}
	}

	// Create Lambda function
	result, err := m.lambdaClient.CreateFunction(ctx, aws.CreateFunctionInput{
		FunctionName: functionName,
		Runtime:      "provided.al2023",
		Role:         roleArn,
		Handler:      "bootstrap",
		Code:         codeBytes,
		MemoryMB:     int32(input.MemoryMB),
		TimeoutSec:   int32(input.TimeoutSec),
		Architecture: input.Architecture,
		Environment:  input.Environment,
		Description:  input.Description,
	})
	if err != nil {
		return "", err
	}

	// Set concurrency if specified
	if input.Concurrency > 0 {
		if err := m.lambdaClient.PutFunctionConcurrency(ctx, functionName, int32(input.Concurrency)); err != nil {
			return "", err
		}
	}

	return *result.FunctionArn, nil
}

// createAPIGateway creates an API Gateway REST API
func (m *Manager) createAPIGateway(ctx context.Context, apiName, lambdaArn, functionName, stage, region string) (string, string, error) {
	// Get AWS account ID from current credentials
	accountID, err := aws.GetAccountID(ctx, m.cfg)
	if err != nil {
		return "", "", fmt.Errorf("failed to get AWS account ID: %w", err)
	}

	// Create REST API
	apiOutput, err := m.agClient.CreateRestAPI(ctx, apiName, "gimage API Gateway")
	if err != nil {
		return "", "", err
	}

	apiID := apiOutput.APIID
	rootID := apiOutput.RootID

	// Create proxy resource
	proxyID, err := m.agClient.CreateProxyResource(ctx, apiID, rootID)
	if err != nil {
		return "", "", err
	}

	// Create ANY method on proxy resource (with API key required)
	if err := m.agClient.CreateMethod(ctx, apiID, proxyID, "ANY", true); err != nil {
		return "", "", err
	}

	// Create ANY method on root resource
	if err := m.agClient.CreateMethod(ctx, apiID, rootID, "ANY", true); err != nil {
		return "", "", err
	}

	// Create Lambda integration for proxy
	if err := m.agClient.CreateLambdaIntegration(ctx, apiID, proxyID, "ANY", lambdaArn, region); err != nil {
		return "", "", err
	}

	// Create Lambda integration for root
	if err := m.agClient.CreateLambdaIntegration(ctx, apiID, rootID, "ANY", lambdaArn, region); err != nil {
		return "", "", err
	}

	// Add permission for API Gateway to invoke Lambda (using actual account ID from credentials)
	sourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*/*", region, accountID, apiID)
	if err := m.lambdaClient.AddPermission(ctx, functionName, "apigateway-invoke", "apigateway.amazonaws.com", sourceArn); err != nil {
		return "", "", err
	}

	// Deploy API
	_, err = m.agClient.DeployAPI(ctx, apiID, stage, "Initial deployment")
	if err != nil {
		return "", "", err
	}

	// Construct API URL
	apiURL := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", apiID, region, stage)

	return apiID, apiURL, nil
}

// findBootstrapFile searches for a bootstrap file in common locations
func findBootstrapFile() string {
	// Check current directory
	if _, err := os.Stat("bootstrap"); err == nil {
		return "bootstrap"
	}

	// Check bin directory
	if _, err := os.Stat("bin/bootstrap"); err == nil {
		return "bin/bootstrap"
	}

	// Check parent directory (for gimage project)
	if _, err := os.Stat("../bootstrap"); err == nil {
		return "../bootstrap"
	}

	return ""
}

// createDeploymentZip creates a deployment zip from a bootstrap file
func createDeploymentZip(bootstrapPath string) ([]byte, error) {
	// For now, just read the file directly
	// In production, this should create a proper zip file
	return os.ReadFile(bootstrapPath)
}

// Destroy removes a deployment and all associated resources
func (m *Manager) Destroy(ctx context.Context, deploymentID string) error {
	// Load deployments
	if err := m.deploymentMgr.Load(); err != nil {
		return fmt.Errorf("failed to load deployments: %w", err)
	}

	// Get deployment
	deployment, err := m.deploymentMgr.Get(deploymentID)
	if err != nil {
		return err
	}

	fmt.Printf("Destroying deployment %s...\n", deploymentID)

	// Delete API Gateway
	if deployment.APIGatewayID != "" {
		fmt.Printf("  [1/4] Deleting API Gateway...\n")
		if err := m.agClient.DeleteRestAPI(ctx, deployment.APIGatewayID); err != nil {
			fmt.Printf("    Warning: Failed to delete API Gateway: %v\n", err)
		}
	}

	// Delete Lambda function
	if deployment.FunctionName != "" {
		fmt.Printf("  [2/4] Deleting Lambda function...\n")
		if err := m.lambdaClient.DeleteFunction(ctx, deployment.FunctionName); err != nil {
			fmt.Printf("    Warning: Failed to delete Lambda function: %v\n", err)
		}
	}

	// Delete S3 bucket
	if deployment.S3Bucket != "" {
		fmt.Printf("  [3/4] Emptying and deleting S3 bucket...\n")
		if err := m.s3Client.EmptyBucket(ctx, deployment.S3Bucket); err != nil {
			fmt.Printf("    Warning: Failed to empty S3 bucket: %v\n", err)
		}
		if err := m.s3Client.DeleteBucket(ctx, deployment.S3Bucket); err != nil {
			fmt.Printf("    Warning: Failed to delete S3 bucket: %v\n", err)
		}
	}

	// Delete IAM role
	if deployment.IAMRoleARN != "" {
		roleName := filepath.Base(deployment.IAMRoleARN)
		fmt.Printf("  [4/4] Deleting IAM role...\n")
		if err := m.iamClient.DeleteRole(ctx, roleName); err != nil {
			fmt.Printf("    Warning: Failed to delete IAM role: %v\n", err)
		}
	}

	// Remove from registry
	if err := m.deploymentMgr.Delete(deploymentID); err != nil {
		return fmt.Errorf("failed to remove deployment from registry: %w", err)
	}

	fmt.Printf("\n✓ Deployment %s destroyed successfully!\n", deploymentID)
	return nil
}
