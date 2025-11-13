package workos

import "fmt"

// AuthorizationPendingError indicates the user hasn't completed authentication yet.
// The client should continue polling.
type AuthorizationPendingError struct{}

func (e *AuthorizationPendingError) Error() string {
	return "authorization pending: user has not completed authentication"
}

// SlowDownError indicates the client is polling too frequently.
// The client should increase the polling interval.
type SlowDownError struct{}

func (e *SlowDownError) Error() string {
	return "slow down: polling too frequently"
}

// ExpiredTokenError indicates the device code has expired.
// The client should restart the device authorization flow.
type ExpiredTokenError struct{}

func (e *ExpiredTokenError) Error() string {
	return "expired token: device code has expired"
}

// AccessDeniedError indicates the user denied authorization.
type AccessDeniedError struct{}

func (e *AccessDeniedError) Error() string {
	return "access denied: user denied authorization"
}

// UnknownError represents an error from WorkOS that we don't have a specific type for.
// This is the fallback for unrecognized error codes.
type UnknownError struct {
	Code        string
	Description string
	StatusCode  int
}

func (e *UnknownError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("WorkOS error %s: %s (status %d)", e.Code, e.Description, e.StatusCode)
	}
	return fmt.Sprintf("WorkOS error %s (status %d)", e.Code, e.StatusCode)
}
