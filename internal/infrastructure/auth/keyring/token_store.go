package keyring

import "github.com/usetero/cli/internal/domains/identity"

// TokenStore maps keyring key-value storage into identity.TokenStore.
type TokenStore struct {
	store *Store
}

// NewTokenStore creates a typed token store backed by keyring storage.
func NewTokenStore(store *Store) *TokenStore {
	if store == nil {
		panic("keyring token store requires store")
	}
	return &TokenStore{store: store}
}

func (s *TokenStore) AccessToken() (identity.AccessToken, error) {
	value, err := s.store.Get(KeyAccessToken)
	return identity.AccessToken(value), err
}

func (s *TokenStore) RefreshToken() (identity.RefreshToken, error) {
	value, err := s.store.Get(KeyRefreshToken)
	return identity.RefreshToken(value), err
}

func (s *TokenStore) SetTokens(tokens identity.Tokens) error {
	if err := s.store.Set(KeyAccessToken, string(tokens.AccessToken)); err != nil {
		return err
	}
	if err := s.store.Set(KeyRefreshToken, string(tokens.RefreshToken)); err != nil {
		return err
	}
	return nil
}

func (s *TokenStore) ClearTokens() error {
	if err := s.store.Delete(KeyAccessToken); err != nil {
		return err
	}
	if err := s.store.Delete(KeyRefreshToken); err != nil {
		return err
	}
	return nil
}
