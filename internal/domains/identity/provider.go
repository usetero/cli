package identity

import (
	"context"

	"github.com/usetero/cli/internal/infrastructure/auth/workos"
)

// ProviderOrgID is the provider-scoped organization identifier.
type ProviderOrgID string

// Provider defines the auth provider behavior needed by identity service.
type Provider interface {
	StartDeviceAuth(ctx context.Context) (DeviceFlow, error)
	PollAuthentication(ctx context.Context, deviceCode string) (Tokens, User, error)
	Refresh(ctx context.Context, refreshToken RefreshToken, providerOrgID ProviderOrgID) (Tokens, error)
}

// WorkOSProvider maps WorkOS client types/errors into identity types/errors.
type WorkOSProvider struct {
	client interface {
		StartDeviceAuthorization(ctx context.Context) (workos.DeviceAuthorization, error)
		PollDeviceAuthorization(ctx context.Context, deviceCode string) (workos.DeviceAuthenticationResult, error)
		RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (workos.RefreshResult, error)
	}
}

// NewWorkOSProvider creates an identity provider backed by WorkOS.
func NewWorkOSProvider(client interface {
	StartDeviceAuthorization(ctx context.Context) (workos.DeviceAuthorization, error)
	PollDeviceAuthorization(ctx context.Context, deviceCode string) (workos.DeviceAuthenticationResult, error)
	RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (workos.RefreshResult, error)
}) *WorkOSProvider {
	if client == nil {
		panic("identity workos provider requires client")
	}
	return &WorkOSProvider{client: client}
}
