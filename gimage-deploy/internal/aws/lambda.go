package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaClient wraps AWS Lambda operations
type LambdaClient struct {
	client *lambda.Client
}

// NewLambdaClient creates a new Lambda client
func NewLambdaClient(cfg aws.Config) *LambdaClient {
	return &LambdaClient{
		client: lambda.NewFromConfig(cfg),
	}
}

// CreateFunctionInput contains parameters for creating a Lambda function
type CreateFunctionInput struct {
	FunctionName string
	Runtime      string
	Role         string
	Handler      string
	Code         []byte
	MemoryMB     int32
	TimeoutSec   int32
	Architecture string
	Environment  map[string]string
	Description  string
}

// CreateFunction creates a new Lambda function
func (lc *LambdaClient) CreateFunction(ctx context.Context, input CreateFunctionInput) (*lambda.CreateFunctionOutput, error) {
	// Convert environment variables
	var envVars *lambdaTypes.Environment
	if len(input.Environment) > 0 {
		variables := make(map[string]string)
		for k, v := range input.Environment {
			variables[k] = v
		}
		envVars = &lambdaTypes.Environment{
			Variables: variables,
		}
	}

	// Determine architecture
	var architectures []lambdaTypes.Architecture
	if input.Architecture == "arm64" {
		architectures = []lambdaTypes.Architecture{lambdaTypes.ArchitectureArm64}
	} else {
		architectures = []lambdaTypes.Architecture{lambdaTypes.ArchitectureX8664}
	}

	createInput := &lambda.CreateFunctionInput{
		FunctionName: aws.String(input.FunctionName),
		Runtime:      lambdaTypes.Runtime(input.Runtime),
		Role:         aws.String(input.Role),
		Handler:      aws.String(input.Handler),
		Code: &lambdaTypes.FunctionCode{
			ZipFile: input.Code,
		},
		MemorySize:    aws.Int32(input.MemoryMB),
		Timeout:       aws.Int32(input.TimeoutSec),
		Architectures: architectures,
		Environment:   envVars,
		Description:   aws.String(input.Description),
	}

	result, err := lc.client.CreateFunction(ctx, createInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda function: %w", err)
	}

	return result, nil
}

// GetFunction retrieves information about a Lambda function
func (lc *LambdaClient) GetFunction(ctx context.Context, functionName string) (*lambda.GetFunctionOutput, error) {
	result, err := lc.client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda function: %w", err)
	}
	return result, nil
}

// UpdateFunctionCode updates the code of a Lambda function
func (lc *LambdaClient) UpdateFunctionCode(ctx context.Context, functionName string, code []byte) error {
	_, err := lc.client.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		ZipFile:      code,
	})
	if err != nil {
		return fmt.Errorf("failed to update Lambda function code: %w", err)
	}
	return nil
}

// UpdateFunctionConfiguration updates the configuration of a Lambda function
func (lc *LambdaClient) UpdateFunctionConfiguration(ctx context.Context, functionName string, memoryMB, timeoutSec int32, env map[string]string) error {
	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		MemorySize:   aws.Int32(memoryMB),
		Timeout:      aws.Int32(timeoutSec),
	}

	if len(env) > 0 {
		input.Environment = &lambdaTypes.Environment{
			Variables: env,
		}
	}

	_, err := lc.client.UpdateFunctionConfiguration(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update Lambda function configuration: %w", err)
	}
	return nil
}

// DeleteFunction deletes a Lambda function
func (lc *LambdaClient) DeleteFunction(ctx context.Context, functionName string) error {
	_, err := lc.client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete Lambda function: %w", err)
	}
	return nil
}

// InvokeFunction invokes a Lambda function
func (lc *LambdaClient) InvokeFunction(ctx context.Context, functionName string, payload []byte) ([]byte, error) {
	result, err := lc.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda function: %w", err)
	}
	return result.Payload, nil
}

// PutFunctionConcurrency sets reserved concurrent executions for a function
func (lc *LambdaClient) PutFunctionConcurrency(ctx context.Context, functionName string, concurrency int32) error {
	_, err := lc.client.PutFunctionConcurrency(ctx, &lambda.PutFunctionConcurrencyInput{
		FunctionName:                 aws.String(functionName),
		ReservedConcurrentExecutions: aws.Int32(concurrency),
	})
	if err != nil {
		return fmt.Errorf("failed to set function concurrency: %w", err)
	}
	return nil
}

// AddPermission adds permission to invoke the Lambda function
func (lc *LambdaClient) AddPermission(ctx context.Context, functionName, statementID, principal, sourceArn string) error {
	_, err := lc.client.AddPermission(ctx, &lambda.AddPermissionInput{
		FunctionName: aws.String(functionName),
		StatementId:  aws.String(statementID),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String(principal),
		SourceArn:    aws.String(sourceArn),
	})
	if err != nil {
		return fmt.Errorf("failed to add permission: %w", err)
	}
	return nil
}
