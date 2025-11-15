package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// deploymentIDPattern allows alphanumeric and hyphens
	deploymentIDPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	// validStages defines allowed deployment stages
	validStages = map[string]bool{
		"production": true,
		"prod":       true,
		"staging":    true,
		"stage":      true,
		"development": true,
		"dev":        true,
		"test":       true,
		"testing":    true,
	}

	// validRegions defines common AWS regions (not exhaustive)
	validRegions = map[string]bool{
		"us-east-1":      true,
		"us-east-2":      true,
		"us-west-1":      true,
		"us-west-2":      true,
		"eu-west-1":      true,
		"eu-west-2":      true,
		"eu-central-1":   true,
		"ap-south-1":     true,
		"ap-southeast-1": true,
		"ap-southeast-2": true,
		"ap-northeast-1": true,
	}
)

// ValidateDeploymentID validates a deployment ID
func ValidateDeploymentID(id string) error {
	if id == "" {
		return fmt.Errorf("deployment ID cannot be empty")
	}
	if len(id) > 64 {
		return fmt.Errorf("deployment ID too long (max 64 characters)")
	}
	if !deploymentIDPattern.MatchString(id) {
		return fmt.Errorf("deployment ID must contain only alphanumeric characters and hyphens")
	}
	return nil
}

// ValidateStage validates a deployment stage
func ValidateStage(stage string) error {
	if stage == "" {
		return fmt.Errorf("stage cannot be empty")
	}
	if !validStages[strings.ToLower(stage)] {
		return fmt.Errorf("invalid stage: %s (must be prod, staging, dev, or test)", stage)
	}
	return nil
}

// ValidateRegion validates an AWS region
func ValidateRegion(region string) error {
	if region == "" {
		return fmt.Errorf("region cannot be empty")
	}
	if !validRegions[region] {
		return fmt.Errorf("invalid or unsupported region: %s", region)
	}
	return nil
}

// ValidateMemory validates Lambda memory configuration
func ValidateMemory(memoryMB int) error {
	if memoryMB < 128 || memoryMB > 10240 {
		return fmt.Errorf("memory must be between 128 and 10240 MB")
	}
	// Memory must be in 1 MB increments
	if memoryMB%1 != 0 {
		return fmt.Errorf("memory must be in 1 MB increments")
	}
	return nil
}

// ValidateTimeout validates Lambda timeout configuration
func ValidateTimeout(timeoutSec int) error {
	if timeoutSec < 1 || timeoutSec > 900 {
		return fmt.Errorf("timeout must be between 1 and 900 seconds")
	}
	return nil
}

// ValidateConcurrency validates Lambda concurrency configuration
func ValidateConcurrency(concurrency int) error {
	if concurrency < 0 || concurrency > 1000 {
		return fmt.Errorf("concurrency must be between 0 and 1000")
	}
	return nil
}

// ValidateAPIKeyName validates an API key name
func ValidateAPIKeyName(name string) error {
	if name == "" {
		return fmt.Errorf("API key name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("API key name too long (max 128 characters)")
	}
	return nil
}
