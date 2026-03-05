package keyring

import (
	"errors"
	"fmt"

	keyringlib "github.com/zalando/go-keyring"
)

// Store persists secrets in OS keychain scoped by runtime environment.
type Store struct {
	service string
}

// NewStore creates an environment-scoped keyring store.
func NewStore(env string) (*Store, error) {
	if env == "" {
		return nil, fmt.Errorf("env is required")
	}
	return &Store{service: baseServiceName + ":" + env}, nil
}

// Get returns a secret value for key, or empty string if not found.
func (s *Store) Get(key string) (string, error) {
	return s.get(key)
}

// Set stores a secret value for key.
func (s *Store) Set(key, value string) error {
	return s.set(key, value)
}

// Delete removes a secret value for key.
func (s *Store) Delete(key string) error {
	return s.delete(key)
}

func (s *Store) get(key string) (string, error) {
	value, err := keyringlib.Get(s.service, key)
	if err != nil {
		if errors.Is(err, keyringlib.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (s *Store) set(key, value string) error {
	return keyringlib.Set(s.service, key, value)
}

func (s *Store) delete(key string) error {
	err := keyringlib.Delete(s.service, key)
	if errors.Is(err, keyringlib.ErrNotFound) {
		return nil
	}
	return err
}
