package apikeys

import (
	"context"
	"fmt"
	"time"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/models"
	"github.com/apresai/gimage-deploy/internal/storage"
	awsConfig "github.com/aws/aws-sdk-go-v2/aws"
)

// Manager handles API key operations
type Manager struct {
	cfg       awsConfig.Config
	agClient  *aws.APIGatewayClient
	keyMgr    *storage.APIKeyManager
	deployMgr *storage.DeploymentManager
}

// NewManager creates a new API key manager
func NewManager(cfg awsConfig.Config) *Manager {
	return &Manager{
		cfg:       cfg,
		agClient:  aws.NewAPIGatewayClient(cfg),
		keyMgr:    storage.NewAPIKeyManager(),
		deployMgr: storage.NewDeploymentManager(),
	}
}

// CreateInput contains parameters for creating an API key
type CreateInput struct {
	Name         string
	DeploymentID string
	Description  string
	RateLimit    int32
	BurstLimit   int32
	QuotaLimit   int32
}

// Create creates a new API key for a deployment
func (m *Manager) Create(ctx context.Context, input CreateInput) (*models.APIKey, error) {
	// Load managers
	if err := m.keyMgr.Load(); err != nil {
		return nil, fmt.Errorf("failed to load API keys: %w", err)
	}
	if err := m.deployMgr.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deployments: %w", err)
	}

	// Get deployment
	deployment, err := m.deployMgr.Get(input.DeploymentID)
	if err != nil {
		return nil, fmt.Errorf("deployment not found: %w", err)
	}

	// Check if key name already exists for this deployment
	existingKeys := m.keyMgr.ListByDeployment(input.DeploymentID)
	for _, key := range existingKeys {
		if key.Name == input.Name {
			return nil, fmt.Errorf("API key with name %s already exists for deployment %s", input.Name, input.DeploymentID)
		}
	}

	fmt.Printf("Creating API key %s for deployment %s...\n", input.Name, input.DeploymentID)

	// Step 1: Create API key in API Gateway
	fmt.Printf("  [1/3] Creating API key in API Gateway...\n")
	keyID, keyValue, err := m.agClient.CreateAPIKey(ctx, input.Name, input.Description, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Step 2: Create usage plan
	fmt.Printf("  [2/4] Creating usage plan...\n")
	usagePlanName := fmt.Sprintf("%s-usage-plan", input.Name)
	usagePlanID, err := m.agClient.CreateUsagePlan(ctx, usagePlanName,
		fmt.Sprintf("Usage plan for %s", input.Name),
		input.RateLimit, input.BurstLimit, input.QuotaLimit)
	if err != nil {
		// Cleanup: delete API key
		m.agClient.DeleteAPIKey(ctx, keyID)
		return nil, fmt.Errorf("failed to create usage plan: %w", err)
	}

	// Step 3: Associate usage plan with API Gateway stage
	fmt.Printf("  [3/4] Associating usage plan with API Gateway stage...\n")
	if err := m.agClient.AssociateAPIStageWithUsagePlan(ctx, usagePlanID, deployment.APIGatewayID, deployment.Stage); err != nil {
		// Cleanup
		m.agClient.DeleteAPIKey(ctx, keyID)
		return nil, fmt.Errorf("failed to associate usage plan with stage: %w", err)
	}

	// Step 4: Associate API key with usage plan
	fmt.Printf("  [4/4] Associating API key with usage plan...\n")
	if err := m.agClient.AssociateAPIKeyWithUsagePlan(ctx, usagePlanID, keyID); err != nil {
		// Cleanup
		m.agClient.DeleteAPIKey(ctx, keyID)
		return nil, fmt.Errorf("failed to associate API key with usage plan: %w", err)
	}

	// Save to local storage
	apiKey := &models.APIKey{
		ID:           keyID,
		Name:         input.Name,
		Description:  input.Description,
		KeyValue:     keyValue,
		DeploymentID: input.DeploymentID,
		Status:       models.APIKeyActive,
		UsagePlanID:  usagePlanID,
		RateLimit:    int(input.RateLimit),
		BurstLimit:   int(input.BurstLimit),
		QuotaLimit:   int(input.QuotaLimit),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := m.keyMgr.Add(apiKey); err != nil {
		return nil, fmt.Errorf("failed to save API key: %w", err)
	}

	fmt.Printf("\n✓ API key %s created successfully!\n", input.Name)
	fmt.Printf("  Key ID:    %s\n", keyID)
	fmt.Printf("  Key Value: %s\n", keyValue)
	fmt.Printf("  Rate:      %d req/sec\n", input.RateLimit)
	fmt.Printf("  Quota:     %d req/day\n", input.QuotaLimit)
	fmt.Printf("\nTest with:\n")
	fmt.Printf("  curl %s/health -H \"X-API-Key: %s\"\n", deployment.APIGatewayURL, keyValue)

	return apiKey, nil
}

// Delete removes an API key
func (m *Manager) Delete(ctx context.Context, keyID string) error {
	// Load keys
	if err := m.keyMgr.Load(); err != nil {
		return fmt.Errorf("failed to load API keys: %w", err)
	}

	// Get key
	key, err := m.keyMgr.Get(keyID)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting API key %s...\n", key.Name)

	// Delete from API Gateway
	if err := m.agClient.DeleteAPIKey(ctx, keyID); err != nil {
		fmt.Printf("  Warning: Failed to delete from API Gateway: %v\n", err)
	}

	// Delete from local storage
	if err := m.keyMgr.Delete(keyID); err != nil {
		return fmt.Errorf("failed to remove from storage: %w", err)
	}

	fmt.Printf("✓ API key %s deleted successfully\n", key.Name)
	return nil
}

// Update modifies an API key
func (m *Manager) Update(ctx context.Context, keyID string, enabled bool) error {
	// Load keys
	if err := m.keyMgr.Load(); err != nil {
		return fmt.Errorf("failed to load API keys: %w", err)
	}

	// Get key
	key, err := m.keyMgr.Get(keyID)
	if err != nil {
		return err
	}

	// Update in API Gateway
	if err := m.agClient.UpdateAPIKey(ctx, keyID, enabled); err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	// Update local storage
	if enabled {
		key.Status = models.APIKeyActive
	} else {
		key.Status = models.APIKeyDisabled
	}
	key.UpdatedAt = time.Now()

	if err := m.keyMgr.Update(key); err != nil {
		return fmt.Errorf("failed to update storage: %w", err)
	}

	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	fmt.Printf("✓ API key %s %s successfully\n", key.Name, status)

	return nil
}
