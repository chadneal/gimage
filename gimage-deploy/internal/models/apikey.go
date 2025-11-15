package models

import "time"

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	APIKeyActive   APIKeyStatus = "active"
	APIKeyDisabled APIKeyStatus = "disabled"
	APIKeyExpired  APIKeyStatus = "expired"
)

// APIKey represents an API Gateway API key
type APIKey struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	KeyValue     string                 `json:"key_value"` // encrypted in storage
	DeploymentID string                 `json:"deployment_id"`
	Status       APIKeyStatus           `json:"status"`
	UsagePlanID  string                 `json:"usage_plan_id"`
	RateLimit    int                    `json:"rate_limit"`  // requests per second
	BurstLimit   int                    `json:"burst_limit"` // burst capacity
	QuotaLimit   int                    `json:"quota_limit"` // requests per day
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastUsed     *time.Time             `json:"last_used,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UsageStats represents API key usage statistics
type UsageStats struct {
	APIKeyID       string          `json:"api_key_id"`
	Period         string          `json:"period"` // 1h, 24h, 7d, 30d
	TotalRequests  int64           `json:"total_requests"`
	ErrorCount     int64           `json:"error_count"`
	ThrottleCount  int64           `json:"throttle_count"`
	QuotaUsed      int64           `json:"quota_used"`
	QuotaRemaining int64           `json:"quota_remaining"`
	TopEndpoints   []EndpointStat  `json:"top_endpoints"`
	ErrorRate      float64         `json:"error_rate"`
	AvgLatency     float64         `json:"avg_latency_ms"`
}

// EndpointStat represents statistics for a specific endpoint
type EndpointStat struct {
	Path         string  `json:"path"`
	RequestCount int64   `json:"request_count"`
	Percentage   float64 `json:"percentage"`
}

// APIKeyRegistry represents the local API key registry
type APIKeyRegistry struct {
	Version    string             `json:"version"`
	Encryption string             `json:"encryption"` // aes-256-gcm
	Keys       map[string]*APIKey `json:"keys"`
	LastSync   time.Time          `json:"last_sync"`
}
