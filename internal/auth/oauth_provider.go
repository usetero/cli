package auth

import "context"

// OAuthProvider defines the interface for OAuth device authorization flow.
// This allows Service to work with any OAuth provider (WorkOS, Auth0, etc.).
// Concrete implementation: workos.Client
type OAuthProvider interface {
	AuthorizeDevice(ctx context.Context) (*DeviceAuthResponse, error)
	PollAuthentication(ctx context.Context, deviceCode string) (*AuthenticationResponse, error)
}

// DeviceAuthResponse represents the response from device authorization.
type DeviceAuthResponse struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               int
	Interval                int
}

// AuthenticationResponse represents a successful authentication.
type AuthenticationResponse struct {
	AccessToken  string
	RefreshToken string
	User         User
}

// OAuth error types - provider-agnostic

// AuthorizationPendingError indicates the user hasn't completed authentication yet.
type AuthorizationPendingError struct{}

func (e *AuthorizationPendingError) Error() string {
	return "authorization pending: user has not completed authentication"
}

// SlowDownError indicates the client is polling too frequently.
type SlowDownError struct{}

func (e *SlowDownError) Error() string {
	return "slow down: polling too frequently"
}

// ExpiredTokenError indicates the device code has expired.
type ExpiredTokenError struct{}

func (e *ExpiredTokenError) Error() string {
	return "expired token: device code has expired"
}

// AccessDeniedError indicates the user denied authorization.
type AccessDeniedError struct{}

func (e *AccessDeniedError) Error() string {
	return "access denied: user denied authorization"
}
