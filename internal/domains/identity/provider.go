package identity

import (
	"context"
)

// ProviderOrgID is the provider-scoped organization identifier.
type ProviderOrgID string

// Provider defines the auth provider behavior needed by identity service.
type Provider interface {
	StartDeviceAuth(ctx context.Context) (DeviceFlow, error)
	PollAuthentication(ctx context.Context, deviceCode string) (Tokens, User, error)
	Refresh(ctx context.Context, refreshToken RefreshToken, providerOrgID ProviderOrgID) (Tokens, error)
}
