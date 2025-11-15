package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	agTypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
)

// APIGatewayClient wraps AWS API Gateway operations
type APIGatewayClient struct {
	client *apigateway.Client
}

// NewAPIGatewayClient creates a new API Gateway client
func NewAPIGatewayClient(cfg aws.Config) *APIGatewayClient {
	return &APIGatewayClient{
		client: apigateway.NewFromConfig(cfg),
	}
}

// CreateRestAPIOutput contains the result of creating a REST API
type CreateRestAPIOutput struct {
	APIID   string
	RootID  string
	APIURL  string
}

// CreateRestAPI creates a new REST API
func (agc *APIGatewayClient) CreateRestAPI(ctx context.Context, name, description string) (*CreateRestAPIOutput, error) {
	// Create REST API
	createOutput, err := agc.client.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
		Name:        aws.String(name),
		Description: aws.String(description),
		EndpointConfiguration: &agTypes.EndpointConfiguration{
			Types: []agTypes.EndpointType{agTypes.EndpointTypeRegional},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create REST API: %w", err)
	}

	apiID := *createOutput.Id

	// Get root resource
	resourcesOutput, err := agc.client.GetResources(ctx, &apigateway.GetResourcesInput{
		RestApiId: aws.String(apiID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	var rootID string
	for _, resource := range resourcesOutput.Items {
		if resource.Path != nil && *resource.Path == "/" {
			rootID = *resource.Id
			break
		}
	}

	return &CreateRestAPIOutput{
		APIID:  apiID,
		RootID: rootID,
		APIURL: "", // Will be set after deployment
	}, nil
}

// CreateProxyResource creates a proxy resource that forwards all requests to Lambda
func (agc *APIGatewayClient) CreateProxyResource(ctx context.Context, apiID, rootID string) (string, error) {
	// Create {proxy+} resource
	resourceOutput, err := agc.client.CreateResource(ctx, &apigateway.CreateResourceInput{
		RestApiId: aws.String(apiID),
		ParentId:  aws.String(rootID),
		PathPart:  aws.String("{proxy+}"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create proxy resource: %w", err)
	}

	return *resourceOutput.Id, nil
}

// CreateMethod creates a method for a resource
func (agc *APIGatewayClient) CreateMethod(ctx context.Context, apiID, resourceID, httpMethod string, apiKeyRequired bool) error {
	_, err := agc.client.PutMethod(ctx, &apigateway.PutMethodInput{
		RestApiId:      aws.String(apiID),
		ResourceId:     aws.String(resourceID),
		HttpMethod:     aws.String(httpMethod),
		AuthorizationType: aws.String("NONE"),
		ApiKeyRequired: apiKeyRequired,
	})
	if err != nil {
		return fmt.Errorf("failed to create method: %w", err)
	}

	return nil
}

// CreateLambdaIntegration creates a Lambda integration for a method
func (agc *APIGatewayClient) CreateLambdaIntegration(ctx context.Context, apiID, resourceID, httpMethod, lambdaArn, region string) error {
	uri := fmt.Sprintf("arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations", region, lambdaArn)

	_, err := agc.client.PutIntegration(ctx, &apigateway.PutIntegrationInput{
		RestApiId:             aws.String(apiID),
		ResourceId:            aws.String(resourceID),
		HttpMethod:            aws.String(httpMethod),
		Type:                  agTypes.IntegrationTypeAwsProxy,
		IntegrationHttpMethod: aws.String("POST"),
		Uri:                   aws.String(uri),
	})
	if err != nil {
		return fmt.Errorf("failed to create Lambda integration: %w", err)
	}

	return nil
}

// DeployAPI creates a deployment and stage
func (agc *APIGatewayClient) DeployAPI(ctx context.Context, apiID, stageName, description string) (string, error) {
	// Create deployment
	deployOutput, err := agc.client.CreateDeployment(ctx, &apigateway.CreateDeploymentInput{
		RestApiId:   aws.String(apiID),
		StageName:   aws.String(stageName),
		Description: aws.String(description),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create deployment: %w", err)
	}

	// Get stage to retrieve URL
	stageOutput, err := agc.client.GetStage(ctx, &apigateway.GetStageInput{
		RestApiId: aws.String(apiID),
		StageName: aws.String(stageName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get stage: %w", err)
	}

	_ = deployOutput
	_ = stageOutput

	// Construct API URL (format: https://{api-id}.execute-api.{region}.amazonaws.com/{stage})
	// Region will be injected by caller
	return apiID, nil
}

// DeleteRestAPI deletes a REST API
func (agc *APIGatewayClient) DeleteRestAPI(ctx context.Context, apiID string) error {
	_, err := agc.client.DeleteRestApi(ctx, &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(apiID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete REST API: %w", err)
	}
	return nil
}

// CreateAPIKey creates an API key
func (agc *APIGatewayClient) CreateAPIKey(ctx context.Context, name, description string, enabled bool) (string, string, error) {
	createOutput, err := agc.client.CreateApiKey(ctx, &apigateway.CreateApiKeyInput{
		Name:        aws.String(name),
		Description: aws.String(description),
		Enabled:     enabled,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create API key: %w", err)
	}

	return *createOutput.Id, *createOutput.Value, nil
}

// CreateUsagePlan creates a usage plan with quotas and throttling
func (agc *APIGatewayClient) CreateUsagePlan(ctx context.Context, name, description string, rateLimit, burstLimit, quotaLimit int32) (string, error) {
	createOutput, err := agc.client.CreateUsagePlan(ctx, &apigateway.CreateUsagePlanInput{
		Name:        aws.String(name),
		Description: aws.String(description),
		Throttle: &agTypes.ThrottleSettings{
			RateLimit:  float64(rateLimit),
			BurstLimit: int32(burstLimit),
		},
		Quota: &agTypes.QuotaSettings{
			Limit:  int32(quotaLimit),
			Period: agTypes.QuotaPeriodTypeDay,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create usage plan: %w", err)
	}

	return *createOutput.Id, nil
}

// AssociateAPIStageWithUsagePlan associates an API stage with a usage plan
func (agc *APIGatewayClient) AssociateAPIStageWithUsagePlan(ctx context.Context, usagePlanID, apiID, stageName string) error {
	_, err := agc.client.UpdateUsagePlan(ctx, &apigateway.UpdateUsagePlanInput{
		UsagePlanId: aws.String(usagePlanID),
		PatchOperations: []agTypes.PatchOperation{
			{
				Op:    agTypes.OpAdd,
				Path:  aws.String("/apiStages"),
				Value: aws.String(fmt.Sprintf("%s:%s", apiID, stageName)),
			},
		},
	})
	if err != nil {
		// If already associated, ignore error
		if !contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to associate stage with usage plan: %w", err)
		}
	}

	return nil
}

// AssociateAPIKeyWithUsagePlan associates an API key with a usage plan
func (agc *APIGatewayClient) AssociateAPIKeyWithUsagePlan(ctx context.Context, usagePlanID, keyID string) error {
	_, err := agc.client.CreateUsagePlanKey(ctx, &apigateway.CreateUsagePlanKeyInput{
		UsagePlanId: aws.String(usagePlanID),
		KeyId:       aws.String(keyID),
		KeyType:     aws.String("API_KEY"),
	})
	if err != nil {
		return fmt.Errorf("failed to associate API key with usage plan: %w", err)
	}

	return nil
}

// DeleteAPIKey deletes an API key
func (agc *APIGatewayClient) DeleteAPIKey(ctx context.Context, keyID string) error {
	_, err := agc.client.DeleteApiKey(ctx, &apigateway.DeleteApiKeyInput{
		ApiKey: aws.String(keyID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	return nil
}

// UpdateAPIKey enables or disables an API key
func (agc *APIGatewayClient) UpdateAPIKey(ctx context.Context, keyID string, enabled bool) error {
	var op agTypes.Op
	var value string
	if enabled {
		op = agTypes.OpReplace
		value = "true"
	} else {
		op = agTypes.OpReplace
		value = "false"
	}

	_, err := agc.client.UpdateApiKey(ctx, &apigateway.UpdateApiKeyInput{
		ApiKey: aws.String(keyID),
		PatchOperations: []agTypes.PatchOperation{
			{
				Op:    op,
				Path:  aws.String("/enabled"),
				Value: aws.String(value),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}
	return nil
}

// GetUsage retrieves usage statistics for an API key
func (agc *APIGatewayClient) GetUsage(ctx context.Context, usagePlanID, keyID, startDate, endDate string) (int64, error) {
	usageOutput, err := agc.client.GetUsage(ctx, &apigateway.GetUsageInput{
		UsagePlanId: aws.String(usagePlanID),
		KeyId:       aws.String(keyID),
		StartDate:   aws.String(startDate),
		EndDate:     aws.String(endDate),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get usage: %w", err)
	}

	// Sum all requests across all items
	var totalRequests int64
	for _, items := range usageOutput.Items {
		for _, count := range items {
			for _, requests := range count {
				totalRequests += requests
			}
		}
	}

	return totalRequests, nil
}
