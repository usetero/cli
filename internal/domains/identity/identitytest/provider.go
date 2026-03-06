package identitytest

import (
	"context"

	"github.com/usetero/cli/internal/domains/identity"
)

type Provider struct {
	StartDeviceAuthFn    func(context.Context) (identity.DeviceFlow, error)
	PollAuthenticationFn func(context.Context, string) (identity.Tokens, identity.User, error)
	RefreshFn            func(context.Context, identity.RefreshToken, identity.ProviderOrgID) (identity.Tokens, error)
}

var _ identity.Provider = (*Provider)(nil)

func (p Provider) StartDeviceAuth(ctx context.Context) (identity.DeviceFlow, error) {
	if p.StartDeviceAuthFn == nil {
		return identity.DeviceFlow{}, nil
	}
	return p.StartDeviceAuthFn(ctx)
}

func (p Provider) PollAuthentication(ctx context.Context, deviceCode string) (identity.Tokens, identity.User, error) {
	if p.PollAuthenticationFn == nil {
		return identity.Tokens{}, identity.User{}, nil
	}
	return p.PollAuthenticationFn(ctx, deviceCode)
}

func (p Provider) Refresh(ctx context.Context, refreshToken identity.RefreshToken, providerOrgID identity.ProviderOrgID) (identity.Tokens, error) {
	if p.RefreshFn == nil {
		return identity.Tokens{}, nil
	}
	return p.RefreshFn(ctx, refreshToken, providerOrgID)
}
