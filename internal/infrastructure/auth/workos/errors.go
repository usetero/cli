package workos

import (
	"errors"
	"fmt"
)

var (
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrSlowDown             = errors.New("slow down")
	ErrExpiredToken         = errors.New("expired token")
	ErrAccessDenied         = errors.New("access denied")
	ErrInvalidGrant         = errors.New("invalid grant")
)

// APIError describes a non-success non-OAuth response from WorkOS.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("workos api error (status %d): %s", e.StatusCode, e.Body)
}

// OAuthError represents an OAuth protocol error from WorkOS.
type OAuthError struct {
	Code        string
	Description string
}

func (e *OAuthError) Error() string {
	if e.Description == "" {
		return fmt.Sprintf("workos oauth error: %s", e.Code)
	}
	return fmt.Sprintf("workos oauth error: %s (%s)", e.Code, e.Description)
}

func (e *OAuthError) Unwrap() error {
	switch e.Code {
	case "authorization_pending":
		return ErrAuthorizationPending
	case "slow_down":
		return ErrSlowDown
	case "expired_token":
		return ErrExpiredToken
	case "access_denied":
		return ErrAccessDenied
	case "invalid_grant":
		return ErrInvalidGrant
	default:
		return nil
	}
}
