package preferences

// Store defines a generic key-value store with persistence.
// This allows services to define domain concepts while keeping the storage implementation generic.
// Concrete implementations: *config.Config (YAML file storage)
type Store interface {
	// Get retrieves a string value by key
	Get(key string) string

	// Set stores a string value by key
	Set(key string, value string)

	// GetBool retrieves a boolean value by key
	GetBool(key string) bool

	// SetBool stores a boolean value by key
	SetBool(key string, value bool)

	// GetList retrieves a list of strings by key
	GetList(key string) []string

	// SetList stores a list of strings by key
	SetList(key string, values []string)

	// Save persists all changes to storage
	Save() error
}
