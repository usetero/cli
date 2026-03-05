package identity

import "time"

// AccessToken is an OAuth access token.
type AccessToken string

// RefreshToken is an OAuth refresh token.
type RefreshToken string

// DeviceFlow contains user-facing and internal values for OAuth device auth.
type DeviceFlow struct {
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	DeviceCode              string
	ExpiresIn               time.Duration
	Interval                time.Duration
}

// Tokens holds OAuth tokens returned by the provider.
type Tokens struct {
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

// User is the authenticated WorkOS user payload.
type User struct {
	ID            string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}
