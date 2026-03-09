package keyring

// Store persists secrets in an environment-scoped backend.
type Store struct {
	backend backend
}

// NewStore creates an environment-scoped secret store.
func NewStore(env string) (*Store, error) {
	if env == "" {
		panic("keyring store requires env")
	}
	backend, err := resolveBackend(env)
	if err != nil {
		return nil, err
	}
	return &Store{backend: backend}, nil
}

// Get returns a secret value for key, or empty string if not found.
func (s *Store) Get(key string) (string, error) {
	return s.backend.Get(key)
}

// Set stores a secret value for key.
func (s *Store) Set(key, value string) error {
	return s.backend.Set(key, value)
}

// Delete removes a secret value for key.
func (s *Store) Delete(key string) error {
	return s.backend.Delete(key)
}
