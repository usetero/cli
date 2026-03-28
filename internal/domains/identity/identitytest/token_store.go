package identitytest

import "github.com/usetero/cli/internal/domains/identity"

type TokenStore struct {
	AccessTokenValue  identity.AccessToken
	RefreshTokenValue identity.RefreshToken
	SetTokensFn       func(identity.Tokens) error
	ClearTokensFn     func() error
}

var _ identity.TokenStore = (*TokenStore)(nil)

func NewTokenStore() *TokenStore {
	return &TokenStore{}
}

func (s *TokenStore) AccessToken() (identity.AccessToken, error) {
	return s.AccessTokenValue, nil
}

func (s *TokenStore) RefreshToken() (identity.RefreshToken, error) {
	return s.RefreshTokenValue, nil
}

func (s *TokenStore) SetTokens(tokens identity.Tokens) error {
	if s.SetTokensFn != nil {
		return s.SetTokensFn(tokens)
	}
	s.AccessTokenValue = tokens.AccessToken
	s.RefreshTokenValue = tokens.RefreshToken
	return nil
}

func (s *TokenStore) ClearTokens() error {
	if s.ClearTokensFn != nil {
		return s.ClearTokensFn()
	}
	s.AccessTokenValue = ""
	s.RefreshTokenValue = ""
	return nil
}
