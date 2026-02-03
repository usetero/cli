package authtest

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// MockAuth implements auth.Auth for testing.
type MockAuth struct {
	StartDeviceAuthFunc                 func(ctx context.Context) (*auth.DeviceAuth, error)
	WaitForAuthFunc                     func(ctx context.Context, deviceCode string, interval time.Duration) (*auth.Result, error)
	IsAuthenticatedFunc                 func() bool
	GetAccessTokenFunc                  func(ctx context.Context) (string, error)
	ClearTokensFunc                     func() error
	RefreshTokenWithoutOrganizationFunc func(ctx context.Context) (string, error)
	RefreshTokenWithOrganizationFunc    func(ctx context.Context, workosOrgID domain.WorkosOrganizationID) (string, error)
}

func (m *MockAuth) StartDeviceAuth(ctx context.Context) (*auth.DeviceAuth, error) {
	if m.StartDeviceAuthFunc != nil {
		return m.StartDeviceAuthFunc(ctx)
	}
	return nil, nil
}

func (m *MockAuth) WaitForAuth(ctx context.Context, deviceCode string, interval time.Duration) (*auth.Result, error) {
	if m.WaitForAuthFunc != nil {
		return m.WaitForAuthFunc(ctx, deviceCode, interval)
	}
	return nil, nil
}

func (m *MockAuth) IsAuthenticated() bool {
	if m.IsAuthenticatedFunc != nil {
		return m.IsAuthenticatedFunc()
	}
	return false
}

func (m *MockAuth) GetAccessToken(ctx context.Context) (string, error) {
	if m.GetAccessTokenFunc != nil {
		return m.GetAccessTokenFunc(ctx)
	}
	return "", nil
}

func (m *MockAuth) ClearTokens() error {
	if m.ClearTokensFunc != nil {
		return m.ClearTokensFunc()
	}
	return nil
}

func (m *MockAuth) RefreshTokenWithoutOrganization(ctx context.Context) (string, error) {
	if m.RefreshTokenWithoutOrganizationFunc != nil {
		return m.RefreshTokenWithoutOrganizationFunc(ctx)
	}
	return "", nil
}

func (m *MockAuth) RefreshTokenWithOrganization(ctx context.Context, workosOrgID domain.WorkosOrganizationID) (string, error) {
	if m.RefreshTokenWithOrganizationFunc != nil {
		return m.RefreshTokenWithOrganizationFunc(ctx, workosOrgID)
	}
	return "", nil
}
