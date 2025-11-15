package storage

import (
	"os"
	"testing"
	"time"

	"github.com/apresai/gimage-deploy/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentManager(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dm := NewDeploymentManager()

	// Test initial state
	deployments := dm.List()
	assert.Empty(t, deployments)

	// Test add
	deployment := &models.Deployment{
		ID:            "test-001",
		Stage:         "test",
		Region:        "us-east-1",
		FunctionName:  "test-function",
		FunctionARN:   "arn:aws:lambda:us-east-1:123456:function:test-function",
		APIGatewayID:  "abc123",
		APIGatewayURL: "https://abc123.execute-api.us-east-1.amazonaws.com/test",
		S3Bucket:      "test-bucket",
		IAMRoleARN:    "arn:aws:iam::123456:role/test-role",
		Status:        models.StatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := dm.Add(deployment)
	assert.NoError(t, err)

	// Test list
	deployments = dm.List()
	assert.Len(t, deployments, 1)
	assert.Equal(t, "test-001", deployments[0].ID)

	// Test get
	retrieved, err := dm.Get("test-001")
	assert.NoError(t, err)
	assert.Equal(t, deployment.ID, retrieved.ID)
	assert.Equal(t, deployment.FunctionName, retrieved.FunctionName)

	// Test exists
	assert.True(t, dm.Exists("test-001"))
	assert.False(t, dm.Exists("nonexistent"))

	// Test update
	deployment.Status = models.StatusFailed
	err = dm.Update(deployment)
	assert.NoError(t, err)

	retrieved, _ = dm.Get("test-001")
	assert.Equal(t, models.StatusFailed, retrieved.Status)

	// Test delete
	err = dm.Delete("test-001")
	assert.NoError(t, err)

	deployments = dm.List()
	assert.Empty(t, deployments)

	// Test error cases
	err = dm.Add(&models.Deployment{}) // No ID
	assert.Error(t, err)

	err = dm.Update(&models.Deployment{ID: "nonexistent"})
	assert.Error(t, err)

	_, err = dm.Get("nonexistent")
	assert.Error(t, err)

	err = dm.Delete("nonexistent")
	assert.Error(t, err)
}
