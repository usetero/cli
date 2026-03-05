package workos

import "time"

// DeviceAuthorization contains user-facing and internal values for device auth.
type DeviceAuthorization struct {
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	DeviceCode              string
	ExpiresIn               time.Duration
	Interval                time.Duration
}

// DeviceAuthenticationResult is returned when device auth succeeds.
type DeviceAuthenticationResult struct {
	AccessToken  string
	RefreshToken string
	User         AuthenticatedUser
}

// AuthenticatedUser is the WorkOS user payload returned during auth.
type AuthenticatedUser struct {
	ID            string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}

type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type authenticateResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
	} `json:"user"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
