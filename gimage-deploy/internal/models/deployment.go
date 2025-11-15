package models

import "time"

// DeploymentStatus represents the current status of a deployment
type DeploymentStatus string

const (
	StatusActive   DeploymentStatus = "active"
	StatusUpdating DeploymentStatus = "updating"
	StatusFailed   DeploymentStatus = "failed"
	StatusInactive DeploymentStatus = "inactive"
	StatusDeleting DeploymentStatus = "deleting"
)

// Deployment represents a Lambda deployment
type Deployment struct {
	ID              string                 `json:"id"`
	Stage           string                 `json:"stage"` // prod, staging, dev, test
	Region          string                 `json:"region"`
	FunctionName    string                 `json:"function_name"`
	FunctionARN     string                 `json:"function_arn"`
	APIGatewayID    string                 `json:"api_gateway_id"`
	APIGatewayURL   string                 `json:"api_gateway_url"`
	S3Bucket        string                 `json:"s3_bucket"`
	IAMRoleARN      string                 `json:"iam_role_arn"`
	Version         string                 `json:"version"` // gimage version
	Status          DeploymentStatus       `json:"status"`
	Health          HealthStatus           `json:"health"`
	Configuration   LambdaConfiguration    `json:"configuration"`
	EnvironmentVars map[string]string      `json:"environment_vars,omitempty"`
	Tags            map[string]string      `json:"tags,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastHealthCheck time.Time              `json:"last_health_check,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus represents the health of a deployment
type HealthStatus struct {
	IsHealthy    bool              `json:"is_healthy"`
	Score        int               `json:"score"` // 0-100
	LastChecked  time.Time         `json:"last_checked"`
	LambdaStatus string            `json:"lambda_status"`
	APIStatus    string            `json:"api_status"`
	S3Status     string            `json:"s3_status"`
	AIProviders  map[string]string `json:"ai_providers"` // gemini: ok, vertex: failed
	ErrorMessage string            `json:"error_message,omitempty"`
}

// LambdaConfiguration represents Lambda function configuration
type LambdaConfiguration struct {
	MemoryMB       int    `json:"memory_mb"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Concurrency    int    `json:"concurrency"`
	Architecture   string `json:"architecture"` // arm64 or x86_64
	Runtime        string `json:"runtime"`      // provided.al2023
	Handler        string `json:"handler"`      // bootstrap (custom runtime)
}

// DeploymentRegistry represents the local deployment registry
type DeploymentRegistry struct {
	Version     string                 `json:"version"`
	Deployments map[string]*Deployment `json:"deployments"`
	LastSync    time.Time              `json:"last_sync"`
}
