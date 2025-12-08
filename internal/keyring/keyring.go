package keyring

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const baseServiceName = "tero-cli"

// Keyring provides secure storage for sensitive data using the system keyring.
// It implements the app.SecureStorage interface using generic key-value operations.
// On macOS it uses Keychain, on Windows it uses Credential Manager, on Linux it uses Secret Service.
type Keyring struct {
	service string
}

// New creates a new keyring.
// If namespace is non-empty, credentials are stored separately (e.g., for non-production environments).
func New(namespace string) *Keyring {
	service := baseServiceName
	if namespace != "" {
		service = baseServiceName + ":" + namespace
	}
	return &Keyring{
		service: service,
	}
}

// Get retrieves a value by key.
// Returns empty string if key doesn't exist.
func (k *Keyring) Get(key string) (string, error) {
	value, err := keyring.Get(k.service, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// Set stores a value by key.
func (k *Keyring) Set(key string, value string) error {
	return keyring.Set(k.service, key, value)
}

// Delete removes a value by key.
func (k *Keyring) Delete(key string) error {
	err := keyring.Delete(k.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil // Not an error if it doesn't exist
	}
	return err
}
