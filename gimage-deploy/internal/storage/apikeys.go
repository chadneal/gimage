package storage

import (
	"fmt"
	"os"
	"time"

	"github.com/apresai/gimage-deploy/internal/models"
	"github.com/apresai/gimage-deploy/pkg/utils"
)

// APIKeyManager handles API key storage with encryption
type APIKeyManager struct {
	registry      *models.APIKeyRegistry
	encryptionKey string
}

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager() *APIKeyManager {
	// Generate encryption key from machine-specific data
	// In production, consider using AWS KMS or a more robust key derivation
	encryptionKey := getEncryptionKey()

	return &APIKeyManager{
		registry: &models.APIKeyRegistry{
			Version:    "1.0.0",
			Encryption: "aes-256-gcm",
			Keys:       make(map[string]*models.APIKey),
			LastSync:   time.Now(),
		},
		encryptionKey: encryptionKey,
	}
}

// getEncryptionKey generates an encryption key from machine-specific data
func getEncryptionKey() string {
	// Use hostname + user as base for encryption key
	// In production, consider using AWS KMS or storing key more securely
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	return fmt.Sprintf("%s-%s-gimage-deploy-key", hostname, user)
}

// Load loads the API key registry from storage
func (am *APIKeyManager) Load() error {
	var registry models.APIKeyRegistry
	if err := LoadJSON(APIKeysFile, &registry); err != nil {
		return err
	}

	// If file doesn't exist, use empty registry
	exists, err := FileExists(APIKeysFile)
	if err != nil {
		return err
	}

	if exists {
		if registry.Keys == nil {
			registry.Keys = make(map[string]*models.APIKey)
		}

		// Decrypt all API key values
		for _, key := range registry.Keys {
			if key.KeyValue != "" {
				decrypted, err := utils.DecryptString(key.KeyValue, am.encryptionKey)
				if err != nil {
					return fmt.Errorf("failed to decrypt API key %s: %w", key.ID, err)
				}
				key.KeyValue = decrypted
			}
		}

		am.registry = &registry
	}

	return nil
}

// Save saves the API key registry to storage (with encryption)
func (am *APIKeyManager) Save() error {
	am.registry.LastSync = time.Now()

	// Create a copy for encryption
	registryCopy := &models.APIKeyRegistry{
		Version:    am.registry.Version,
		Encryption: am.registry.Encryption,
		Keys:       make(map[string]*models.APIKey),
		LastSync:   am.registry.LastSync,
	}

	// Encrypt all API key values before saving
	for id, key := range am.registry.Keys {
		keyCopy := *key
		if keyCopy.KeyValue != "" {
			encrypted, err := utils.EncryptString(keyCopy.KeyValue, am.encryptionKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt API key %s: %w", id, err)
			}
			keyCopy.KeyValue = encrypted
		}
		registryCopy.Keys[id] = &keyCopy
	}

	return SaveJSON(APIKeysFile, registryCopy)
}

// Add adds a new API key to the registry
func (am *APIKeyManager) Add(key *models.APIKey) error {
	if key.ID == "" {
		return fmt.Errorf("API key ID cannot be empty")
	}

	if _, exists := am.registry.Keys[key.ID]; exists {
		return fmt.Errorf("API key with ID %s already exists", key.ID)
	}

	am.registry.Keys[key.ID] = key
	return am.Save()
}

// Update updates an existing API key in the registry
func (am *APIKeyManager) Update(key *models.APIKey) error {
	if key.ID == "" {
		return fmt.Errorf("API key ID cannot be empty")
	}

	if _, exists := am.registry.Keys[key.ID]; !exists {
		return fmt.Errorf("API key with ID %s does not exist", key.ID)
	}

	key.UpdatedAt = time.Now()
	am.registry.Keys[key.ID] = key
	return am.Save()
}

// Delete removes an API key from the registry
func (am *APIKeyManager) Delete(id string) error {
	if _, exists := am.registry.Keys[id]; !exists {
		return fmt.Errorf("API key with ID %s does not exist", id)
	}

	delete(am.registry.Keys, id)
	return am.Save()
}

// Get retrieves an API key by ID
func (am *APIKeyManager) Get(id string) (*models.APIKey, error) {
	key, exists := am.registry.Keys[id]
	if !exists {
		return nil, fmt.Errorf("API key with ID %s not found", id)
	}
	return key, nil
}

// List returns all API keys
func (am *APIKeyManager) List() []*models.APIKey {
	keys := make([]*models.APIKey, 0, len(am.registry.Keys))
	for _, key := range am.registry.Keys {
		keys = append(keys, key)
	}
	return keys
}

// ListByDeployment returns all API keys for a specific deployment
func (am *APIKeyManager) ListByDeployment(deploymentID string) []*models.APIKey {
	keys := make([]*models.APIKey, 0)
	for _, key := range am.registry.Keys {
		if key.DeploymentID == deploymentID {
			keys = append(keys, key)
		}
	}
	return keys
}

// Exists checks if an API key exists
func (am *APIKeyManager) Exists(id string) bool {
	_, exists := am.registry.Keys[id]
	return exists
}
