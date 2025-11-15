package storage

import (
	"fmt"
	"time"

	"github.com/apresai/gimage-deploy/internal/models"
)

// DeploymentManager handles deployment storage
type DeploymentManager struct {
	registry *models.DeploymentRegistry
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager() *DeploymentManager {
	return &DeploymentManager{
		registry: &models.DeploymentRegistry{
			Version:     "1.0.0",
			Deployments: make(map[string]*models.Deployment),
			LastSync:    time.Now(),
		},
	}
}

// Load loads the deployment registry from storage
func (dm *DeploymentManager) Load() error {
	var registry models.DeploymentRegistry
	if err := LoadJSON(DeploymentsFile, &registry); err != nil {
		return err
	}

	// If file doesn't exist, use empty registry
	exists, err := FileExists(DeploymentsFile)
	if err != nil {
		return err
	}

	if exists {
		if registry.Deployments == nil {
			registry.Deployments = make(map[string]*models.Deployment)
		}
		dm.registry = &registry
	}

	return nil
}

// Save saves the deployment registry to storage
func (dm *DeploymentManager) Save() error {
	dm.registry.LastSync = time.Now()
	return SaveJSON(DeploymentsFile, dm.registry)
}

// Add adds a new deployment to the registry
func (dm *DeploymentManager) Add(deployment *models.Deployment) error {
	if deployment.ID == "" {
		return fmt.Errorf("deployment ID cannot be empty")
	}

	if _, exists := dm.registry.Deployments[deployment.ID]; exists {
		return fmt.Errorf("deployment with ID %s already exists", deployment.ID)
	}

	dm.registry.Deployments[deployment.ID] = deployment
	return dm.Save()
}

// Update updates an existing deployment in the registry
func (dm *DeploymentManager) Update(deployment *models.Deployment) error {
	if deployment.ID == "" {
		return fmt.Errorf("deployment ID cannot be empty")
	}

	if _, exists := dm.registry.Deployments[deployment.ID]; !exists {
		return fmt.Errorf("deployment with ID %s does not exist", deployment.ID)
	}

	deployment.UpdatedAt = time.Now()
	dm.registry.Deployments[deployment.ID] = deployment
	return dm.Save()
}

// Delete removes a deployment from the registry
func (dm *DeploymentManager) Delete(id string) error {
	if _, exists := dm.registry.Deployments[id]; !exists {
		return fmt.Errorf("deployment with ID %s does not exist", id)
	}

	delete(dm.registry.Deployments, id)
	return dm.Save()
}

// Get retrieves a deployment by ID
func (dm *DeploymentManager) Get(id string) (*models.Deployment, error) {
	deployment, exists := dm.registry.Deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment with ID %s not found", id)
	}
	return deployment, nil
}

// List returns all deployments
func (dm *DeploymentManager) List() []*models.Deployment {
	deployments := make([]*models.Deployment, 0, len(dm.registry.Deployments))
	for _, deployment := range dm.registry.Deployments {
		deployments = append(deployments, deployment)
	}
	return deployments
}

// Exists checks if a deployment exists
func (dm *DeploymentManager) Exists(id string) bool {
	_, exists := dm.registry.Deployments[id]
	return exists
}
