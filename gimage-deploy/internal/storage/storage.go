package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// StorageDir is the directory where all data is stored
	StorageDir = ".gimage-deploy"

	// ConfigFile is the configuration file name
	ConfigFile = "config.json"

	// DeploymentsFile is the deployments registry file name
	DeploymentsFile = "deployments.json"

	// APIKeysFile is the API keys registry file name
	APIKeysFile = "api_keys.encrypted.json"

	// FilePermissions for storage files (owner read/write only)
	FilePermissions = 0600

	// DirPermissions for storage directory
	DirPermissions = 0700
)

// GetStorageDir returns the absolute path to the storage directory
func GetStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, StorageDir), nil
}

// EnsureStorageDir creates the storage directory if it doesn't exist
func EnsureStorageDir() error {
	dir, err := GetStorageDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, DirPermissions); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Also create cache subdirectory
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(cacheDir, DirPermissions); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	return nil
}

// LoadJSON loads a JSON file from storage
func LoadJSON(filename string, v interface{}) error {
	dir, err := GetStorageDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, not an error
		}
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	return nil
}

// SaveJSON saves a value as JSON to storage
func SaveJSON(filename string, v interface{}) error {
	dir, err := GetStorageDir()
	if err != nil {
		return err
	}

	if err := EnsureStorageDir(); err != nil {
		return err
	}

	path := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", filename, err)
	}

	if err := os.WriteFile(path, data, FilePermissions); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}

// FileExists checks if a file exists in storage
func FileExists(filename string) (bool, error) {
	dir, err := GetStorageDir()
	if err != nil {
		return false, err
	}

	path := filepath.Join(dir, filename)
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
