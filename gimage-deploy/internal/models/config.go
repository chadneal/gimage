package models

// Config represents the local configuration
type Config struct {
	AWSProfile         string                 `json:"aws_profile"`
	DefaultRegion      string                 `json:"default_region"`
	DefaultStage       string                 `json:"default_stage"`
	DefaultMemoryMB    int                    `json:"default_memory_mb"`
	DefaultTimeoutSec  int                    `json:"default_timeout_sec"`
	DefaultConcurrency int                    `json:"default_concurrency"`
	AutoRefreshMetrics bool                   `json:"auto_refresh_metrics"`
	RefreshIntervalSec int                    `json:"refresh_interval_sec"`
	LogRetentionDays   int                    `json:"log_retention_days"`
	EncryptionEnabled  bool                   `json:"encryption_enabled"`
	Preferences        map[string]interface{} `json:"preferences,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		AWSProfile:         "default",
		DefaultRegion:      "us-east-1",
		DefaultStage:       "production",
		DefaultMemoryMB:    512,
		DefaultTimeoutSec:  30,
		DefaultConcurrency: 10,
		AutoRefreshMetrics: true,
		RefreshIntervalSec: 5,
		LogRetentionDays:   7,
		EncryptionEnabled:  true,
		Preferences:        make(map[string]interface{}),
	}
}
