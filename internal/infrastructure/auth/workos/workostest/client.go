package workostest

import (
	"context"

	"github.com/usetero/cli/internal/infrastructure/auth/workos"
)

// Client is a functional mock for WorkOS auth client flows.
type Client struct {
	StartDeviceAuthorizationFn func(ctx context.Context) (workos.DeviceAuthorization, error)
	PollDeviceAuthorizationFn  func(ctx context.Context, deviceCode string) (workos.DeviceAuthenticationResult, error)
	RefreshTokenFn             func(ctx context.Context, refreshToken, workosOrgID string) (workos.RefreshResult, error)
}

var _ interface {
	StartDeviceAuthorization(ctx context.Context) (workos.DeviceAuthorization, error)
	PollDeviceAuthorization(ctx context.Context, deviceCode string) (workos.DeviceAuthenticationResult, error)
	RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (workos.RefreshResult, error)
} = (*Client)(nil)

func (c *Client) StartDeviceAuthorization(ctx context.Context) (workos.DeviceAuthorization, error) {
	if c.StartDeviceAuthorizationFn == nil {
		return workos.DeviceAuthorization{}, nil
	}
	return c.StartDeviceAuthorizationFn(ctx)
}

func (c *Client) PollDeviceAuthorization(ctx context.Context, deviceCode string) (workos.DeviceAuthenticationResult, error) {
	if c.PollDeviceAuthorizationFn == nil {
		return workos.DeviceAuthenticationResult{}, nil
	}
	return c.PollDeviceAuthorizationFn(ctx, deviceCode)
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (workos.RefreshResult, error) {
	if c.RefreshTokenFn == nil {
		return workos.RefreshResult{}, nil
	}
	return c.RefreshTokenFn(ctx, refreshToken, workosOrgID)
}
