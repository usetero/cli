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

// New creates a new keyring scoped to the given environment.
// Migrates credentials from the old service name on first use.
func New(env string) *Keyring {
	k := &Keyring{
		service: baseServiceName + ":" + env,
	}
	k.migrate(env)
	return k
}

// migrate copies credentials from the old service name to the new one.
// Old: "tero-cli" (production), "tero-cli:{host}" (non-production)
// New: "tero-cli:prd", "tero-cli:local", etc.
// Idempotent — skips if new service already has credentials.
func (k *Keyring) migrate(env string) {
	// If we already have credentials, nothing to do.
	if v, _ := keyring.Get(k.service, "access_token"); v != "" {
		return
	}

	// Determine old service name.
	var oldService string
	if env == "prd" {
		oldService = baseServiceName // was "tero-cli" with no suffix
	} else {
		// Can't know the old host — skip migration for non-prd.
		// Developers will just re-login.
		return
	}

	// Copy each key from old to new.
	for _, key := range []string{"access_token", "refresh_token"} {
		val, err := keyring.Get(oldService, key)
		if err != nil || val == "" {
			continue
		}
		_ = keyring.Set(k.service, key, val)
		_ = keyring.Delete(oldService, key)
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
