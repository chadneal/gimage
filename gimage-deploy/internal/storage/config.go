package storage

import (
	"fmt"

	"github.com/apresai/gimage-deploy/internal/models"
)

// ConfigManager handles configuration storage
type ConfigManager struct {
	config *models.Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: models.DefaultConfig(),
	}
}

// Load loads the configuration from storage
func (cm *ConfigManager) Load() error {
	var config models.Config
	if err := LoadJSON(ConfigFile, &config); err != nil {
		return err
	}

	// If config file doesn't exist, use defaults
	exists, err := FileExists(ConfigFile)
	if err != nil {
		return err
	}

	if exists {
		cm.config = &config
	}

	return nil
}

// Save saves the configuration to storage
func (cm *ConfigManager) Save() error {
	return SaveJSON(ConfigFile, cm.config)
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *models.Config {
	return cm.config
}

// Set updates a configuration value
func (cm *ConfigManager) Set(key string, value interface{}) error {
	switch key {
	case "aws_profile":
		if v, ok := value.(string); ok {
			cm.config.AWSProfile = v
		} else {
			return fmt.Errorf("aws_profile must be a string")
		}
	case "default_region":
		if v, ok := value.(string); ok {
			cm.config.DefaultRegion = v
		} else {
			return fmt.Errorf("default_region must be a string")
		}
	case "default_stage":
		if v, ok := value.(string); ok {
			cm.config.DefaultStage = v
		} else {
			return fmt.Errorf("default_stage must be a string")
		}
	case "default_memory_mb":
		if v, ok := value.(int); ok {
			cm.config.DefaultMemoryMB = v
		} else {
			return fmt.Errorf("default_memory_mb must be an integer")
		}
	case "default_timeout_sec":
		if v, ok := value.(int); ok {
			cm.config.DefaultTimeoutSec = v
		} else {
			return fmt.Errorf("default_timeout_sec must be an integer")
		}
	case "default_concurrency":
		if v, ok := value.(int); ok {
			cm.config.DefaultConcurrency = v
		} else {
			return fmt.Errorf("default_concurrency must be an integer")
		}
	case "auto_refresh_metrics":
		if v, ok := value.(bool); ok {
			cm.config.AutoRefreshMetrics = v
		} else {
			return fmt.Errorf("auto_refresh_metrics must be a boolean")
		}
	case "refresh_interval_sec":
		if v, ok := value.(int); ok {
			cm.config.RefreshIntervalSec = v
		} else {
			return fmt.Errorf("refresh_interval_sec must be an integer")
		}
	case "log_retention_days":
		if v, ok := value.(int); ok {
			cm.config.LogRetentionDays = v
		} else {
			return fmt.Errorf("log_retention_days must be an integer")
		}
	case "encryption_enabled":
		if v, ok := value.(bool); ok {
			cm.config.EncryptionEnabled = v
		} else {
			return fmt.Errorf("encryption_enabled must be a boolean")
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return cm.Save()
}

// Reset resets the configuration to defaults
func (cm *ConfigManager) Reset() error {
	cm.config = models.DefaultConfig()
	return cm.Save()
}
