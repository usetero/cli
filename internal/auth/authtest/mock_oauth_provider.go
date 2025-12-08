package authtest

import (
	"context"

	"github.com/usetero/cli/internal/auth"
)

// MockOAuthProvider is a test double for auth.OAuthProvider.
type MockOAuthProvider struct {
	AuthorizeDeviceFunc              func(ctx context.Context) (*auth.DeviceAuthResponse, error)
	PollAuthenticationFunc           func(ctx context.Context, deviceCode string) (*auth.AuthenticationResponse, error)
	RefreshTokenFunc                 func(ctx context.Context, refreshToken string) (*auth.RefreshResponse, error)
	RefreshTokenWithOrganizationFunc func(ctx context.Context, refreshToken, organizationID string) (*auth.RefreshResponse, error)
}

func (m *MockOAuthProvider) AuthorizeDevice(ctx context.Context) (*auth.DeviceAuthResponse, error) {
	if m.AuthorizeDeviceFunc != nil {
		return m.AuthorizeDeviceFunc(ctx)
	}
	return nil, nil
}

func (m *MockOAuthProvider) PollAuthentication(ctx context.Context, deviceCode string) (*auth.AuthenticationResponse, error) {
	if m.PollAuthenticationFunc != nil {
		return m.PollAuthenticationFunc(ctx, deviceCode)
	}
	return nil, nil
}

func (m *MockOAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *MockOAuthProvider) RefreshTokenWithOrganization(ctx context.Context, refreshToken, organizationID string) (*auth.RefreshResponse, error) {
	if m.RefreshTokenWithOrganizationFunc != nil {
		return m.RefreshTokenWithOrganizationFunc(ctx, refreshToken, organizationID)
	}
	return nil, nil
}
