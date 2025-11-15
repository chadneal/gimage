package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// IAMClient wraps AWS IAM operations
type IAMClient struct {
	client *iam.Client
}

// NewIAMClient creates a new IAM client
func NewIAMClient(cfg aws.Config) *IAMClient {
	return &IAMClient{
		client: iam.NewFromConfig(cfg),
	}
}

// CreateLambdaExecutionRole creates an IAM role for Lambda execution
func (ic *IAMClient) CreateLambdaExecutionRole(ctx context.Context, roleName string) (string, error) {
	// Trust policy for Lambda
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"Service": "lambda.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	}

	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal trust policy: %w", err)
	}

	// Create role
	createRoleOutput, err := ic.client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
		Description:              aws.String("Execution role for gimage Lambda function"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create IAM role: %w", err)
	}

	return *createRoleOutput.Role.Arn, nil
}

// AttachLambdaBasicExecutionPolicy attaches basic Lambda execution policy
func (ic *IAMClient) AttachLambdaBasicExecutionPolicy(ctx context.Context, roleName string) error {
	policyArn := "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"

	_, err := ic.client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return fmt.Errorf("failed to attach basic execution policy: %w", err)
	}

	return nil
}

// CreateS3AccessPolicy creates and attaches a policy for S3 access
func (ic *IAMClient) CreateS3AccessPolicy(ctx context.Context, roleName, bucketName string) error {
	policyName := fmt.Sprintf("%s-s3-policy", roleName)

	// S3 access policy
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"s3:PutObject",
					"s3:GetObject",
					"s3:DeleteObject",
					"s3:ListBucket",
				},
				"Resource": []string{
					fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
					fmt.Sprintf("arn:aws:s3:::%s", bucketName),
				},
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal S3 policy: %w", err)
	}

	// Create policy
	createPolicyOutput, err := ic.client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(policyJSON)),
		Description:    aws.String(fmt.Sprintf("S3 access for gimage bucket %s", bucketName)),
	})
	if err != nil {
		return fmt.Errorf("failed to create S3 policy: %w", err)
	}

	// Attach policy to role
	_, err = ic.client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: createPolicyOutput.Policy.Arn,
	})
	if err != nil {
		return fmt.Errorf("failed to attach S3 policy: %w", err)
	}

	return nil
}

// CreateBedrockAccessPolicy creates and attaches a policy for Bedrock access
func (ic *IAMClient) CreateBedrockAccessPolicy(ctx context.Context, roleName string) error {
	policyName := fmt.Sprintf("%s-bedrock-policy", roleName)

	// Bedrock access policy
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"bedrock:InvokeModel",
					"bedrock:InvokeModelWithResponseStream",
				},
				"Resource": "*",
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal Bedrock policy: %w", err)
	}

	// Create policy
	createPolicyOutput, err := ic.client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(policyJSON)),
		Description:    aws.String("Bedrock access for gimage Lambda function"),
	})
	if err != nil {
		return fmt.Errorf("failed to create Bedrock policy: %w", err)
	}

	// Attach policy to role
	_, err = ic.client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: createPolicyOutput.Policy.Arn,
	})
	if err != nil {
		return fmt.Errorf("failed to attach Bedrock policy: %w", err)
	}

	return nil
}

// DeleteRole deletes an IAM role and detaches all policies
func (ic *IAMClient) DeleteRole(ctx context.Context, roleName string) error {
	// List and detach all attached policies
	listPoliciesOutput, err := ic.client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to list attached policies: %w", err)
	}

	for _, policy := range listPoliciesOutput.AttachedPolicies {
		_, err := ic.client.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: policy.PolicyArn,
		})
		if err != nil {
			return fmt.Errorf("failed to detach policy: %w", err)
		}
	}

	// Delete role
	_, err = ic.client.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// RoleExists checks if a role exists
func (ic *IAMClient) RoleExists(ctx context.Context, roleName string) (bool, error) {
	_, err := ic.client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		if isNoSuchEntityError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}
	return true, nil
}

// isNoSuchEntityError checks if the error is a "NoSuchEntity" error
func isNoSuchEntityError(err error) bool {
	return err != nil && (
		err.Error() == "NoSuchEntity" ||
		contains(err.Error(), "NoSuchEntity"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
