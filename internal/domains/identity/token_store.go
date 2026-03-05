package identity

// TokenStore is typed token persistence used by identity service.
type TokenStore interface {
	AccessToken() (AccessToken, error)
	RefreshToken() (RefreshToken, error)
	SetTokens(tokens Tokens) error
	ClearTokens() error
}
