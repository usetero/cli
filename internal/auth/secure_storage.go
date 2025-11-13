package auth

// SecureStorage defines a generic secure key-value storage interface.
// This allows services to define domain concepts (like "access_token") while keeping
// the storage implementation generic (OS keychain, encrypted file, etc.).
// Concrete implementations: keyring.Keyring (OS keychain)
type SecureStorage interface {
	// Get retrieves a value by key
	// Returns empty string if key doesn't exist
	Get(key string) (string, error)

	// Set stores a value by key
	Set(key string, value string) error

	// Delete removes a value by key
	Delete(key string) error
}
