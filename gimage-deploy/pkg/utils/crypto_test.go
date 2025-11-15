package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
	}{
		{"Simple text", "hello world", "secret-key"},
		{"API key", "gimage_k3y_abc123xyz789", "encryption-key"},
		{"Empty string", "", "key"},
		{"Long text", "this is a very long string that should still be encrypted and decrypted correctly", "my-secret-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := EncryptString(tt.plaintext, tt.key)
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.plaintext, encrypted)

			// Decrypt
			decrypted, err := DecryptString(encrypted, tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	plaintext := "secret message"
	correctKey := "correct-key"
	wrongKey := "wrong-key"

	// Encrypt with correct key
	encrypted, err := EncryptString(plaintext, correctKey)
	assert.NoError(t, err)

	// Try to decrypt with wrong key
	_, err = DecryptString(encrypted, wrongKey)
	assert.Error(t, err)
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"Standard API key", "gimage_k3y_abc123xyz789", "gimage_k3y_a***z789"},
		{"Short key", "short", "***"},
		{"Exactly 17 chars", "12345678901234567", "123456789012***4567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskAPIKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
