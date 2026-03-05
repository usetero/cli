package identity

import (
	"github.com/usetero/cli/internal/infrastructure/auth/keyring"
)

// KeyringTokenStore maps keyring key-value storage into identity.TokenStore.
type KeyringTokenStore struct {
	store *keyring.Store
}

// NewKeyringTokenStore creates a typed token store backed by keyring storage.
func NewKeyringTokenStore(store *keyring.Store) *KeyringTokenStore {
	return &KeyringTokenStore{store: store}
}

func (s *KeyringTokenStore) AccessToken() (AccessToken, error) {
	value, err := s.store.Get(keyring.KeyAccessToken)
	return AccessToken(value), err
}

func (s *KeyringTokenStore) RefreshToken() (RefreshToken, error) {
	value, err := s.store.Get(keyring.KeyRefreshToken)
	return RefreshToken(value), err
}

func (s *KeyringTokenStore) SetTokens(tokens Tokens) error {
	if err := s.store.Set(keyring.KeyAccessToken, string(tokens.AccessToken)); err != nil {
		return err
	}
	if err := s.store.Set(keyring.KeyRefreshToken, string(tokens.RefreshToken)); err != nil {
		return err
	}
	return nil
}

func (s *KeyringTokenStore) ClearTokens() error {
	if err := s.store.Delete(keyring.KeyAccessToken); err != nil {
		return err
	}
	if err := s.store.Delete(keyring.KeyRefreshToken); err != nil {
		return err
	}
	return nil
}
