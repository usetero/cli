package identity

import "errors"

var (
	ErrAuthorizationPending  = errors.New("authorization pending")
	ErrSlowDown              = errors.New("slow down")
	ErrExpiredDeviceCode     = errors.New("expired device code")
	ErrAccessDenied          = errors.New("access denied")
	ErrNotAuthenticated      = errors.New("not authenticated")
	ErrProviderNotConfigured = errors.New("auth provider is not configured")
)
