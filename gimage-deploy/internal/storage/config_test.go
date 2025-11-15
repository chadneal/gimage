package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigManager(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cm := NewConfigManager()

	// Test default config
	config := cm.Get()
	assert.Equal(t, "default", config.AWSProfile)
	assert.Equal(t, "us-east-1", config.DefaultRegion)
	assert.Equal(t, "production", config.DefaultStage)
	assert.Equal(t, 512, config.DefaultMemoryMB)
	assert.Equal(t, 30, config.DefaultTimeoutSec)
	assert.Equal(t, 10, config.DefaultConcurrency)

	// Test save
	err := cm.Save()
	assert.NoError(t, err)

	// Verify file exists
	configPath := filepath.Join(tmpDir, StorageDir, ConfigFile)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Test load
	cm2 := NewConfigManager()
	err = cm2.Load()
	assert.NoError(t, err)
	assert.Equal(t, config.AWSProfile, cm2.Get().AWSProfile)

	// Test set
	err = cm.Set("default_region", "us-west-2")
	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", cm.Get().DefaultRegion)

	// Test set with wrong type
	err = cm.Set("default_memory_mb", "not-an-int")
	assert.Error(t, err)

	// Test reset
	err = cm.Reset()
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", cm.Get().DefaultRegion) // Back to default
}
