package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDeploymentID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Valid ID", "prod-001", false},
		{"Valid with numbers", "staging-123", false},
		{"Valid simple", "dev", false},
		{"Empty ID", "", true},
		{"Too long", "this-is-a-very-long-deployment-id-that-exceeds-the-maximum-allowed-length-of-64-characters", true},
		{"Invalid characters", "prod_001", true},
		{"With spaces", "prod 001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeploymentID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStage(t *testing.T) {
	tests := []struct {
		name    string
		stage   string
		wantErr bool
	}{
		{"Production", "production", false},
		{"Prod", "prod", false},
		{"Staging", "staging", false},
		{"Stage", "stage", false},
		{"Development", "development", false},
		{"Dev", "dev", false},
		{"Test", "test", false},
		{"Empty", "", true},
		{"Invalid", "invalid-stage", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStage(tt.stage)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMemory(t *testing.T) {
	tests := []struct {
		name    string
		memory  int
		wantErr bool
	}{
		{"Valid 128", 128, false},
		{"Valid 512", 512, false},
		{"Valid 1024", 1024, false},
		{"Valid 10240", 10240, false},
		{"Too low", 64, true},
		{"Too high", 20000, true},
		{"Zero", 0, true},
		{"Negative", -512, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMemory(tt.memory)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		wantErr bool
	}{
		{"Valid 1", 1, false},
		{"Valid 30", 30, false},
		{"Valid 900", 900, false},
		{"Too low", 0, true},
		{"Too high", 1000, true},
		{"Negative", -30, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeout(tt.timeout)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConcurrency(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		wantErr     bool
	}{
		{"Valid 0", 0, false},
		{"Valid 10", 10, false},
		{"Valid 1000", 1000, false},
		{"Too high", 1001, true},
		{"Negative", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConcurrency(tt.concurrency)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAPIKeyName(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		wantErr bool
	}{
		{"Valid name", "web-app", false},
		{"Valid with numbers", "mobile-app-v2", false},
		{"Valid long", "production-web-application-frontend", false},
		{"Empty", "", true},
		{"Too long", "this-is-an-extremely-long-api-key-name-that-exceeds-the-maximum-allowed-length-of-128-characters-and-should-fail-validation-test-case", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKeyName(tt.keyName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
