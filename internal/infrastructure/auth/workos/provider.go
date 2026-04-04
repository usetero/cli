package workos

import (
	"context"
	"errors"

	"github.com/usetero/cli/internal/domains/identity"
)

// IdentityProvider maps WorkOS client types and errors into identity contracts.
type IdentityProvider struct {
	client interface {
		StartDeviceAuthorization(ctx context.Context) (DeviceAuthorization, error)
		PollDeviceAuthorization(ctx context.Context, deviceCode string) (DeviceAuthenticationResult, error)
		RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (RefreshResult, error)
	}
}

// NewProvider creates an identity.Provider backed by WorkOS.
func NewProvider(client interface {
	StartDeviceAuthorization(ctx context.Context) (DeviceAuthorization, error)
	PollDeviceAuthorization(ctx context.Context, deviceCode string) (DeviceAuthenticationResult, error)
	RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (RefreshResult, error)
}) *IdentityProvider {
	if client == nil {
		panic("workos identity provider requires client")
	}
	return &IdentityProvider{client: client}
}

func (p *IdentityProvider) StartDeviceAuth(ctx context.Context) (identity.DeviceFlow, error) {
	flow, err := p.client.StartDeviceAuthorization(ctx)
	if err != nil {
		return identity.DeviceFlow{}, mapIdentityError(err)
	}
	return identity.DeviceFlow{
		UserCode:                flow.UserCode,
		VerificationURI:         flow.VerificationURI,
		VerificationURIComplete: flow.VerificationURIComplete,
		DeviceCode:              flow.DeviceCode,
		ExpiresIn:               flow.ExpiresIn,
		Interval:                flow.Interval,
	}, nil
}

func (p *IdentityProvider) PollAuthentication(ctx context.Context, deviceCode string) (identity.Tokens, identity.User, error) {
	result, err := p.client.PollDeviceAuthorization(ctx, deviceCode)
	if err != nil {
		return identity.Tokens{}, identity.User{}, mapIdentityError(err)
	}
	return identity.Tokens{
			AccessToken:  identity.AccessToken(result.AccessToken),
			RefreshToken: identity.RefreshToken(result.RefreshToken),
		}, identity.User{
			ID:            result.User.ID,
			Email:         result.User.Email,
			EmailVerified: result.User.EmailVerified,
			FirstName:     result.User.FirstName,
			LastName:      result.User.LastName,
		}, nil
}

func (p *IdentityProvider) Refresh(ctx context.Context, refreshToken identity.RefreshToken, providerOrgID identity.ProviderOrgID) (identity.Tokens, error) {
	result, err := p.client.RefreshToken(ctx, string(refreshToken), string(providerOrgID))
	if err != nil {
		return identity.Tokens{}, mapIdentityError(err)
	}
	return identity.Tokens{
		AccessToken:  identity.AccessToken(result.AccessToken),
		RefreshToken: identity.RefreshToken(result.RefreshToken),
	}, nil
}

func mapIdentityError(err error) error {
	switch {
	case errors.Is(err, ErrAuthorizationPending):
		return identity.ErrAuthorizationPending
	case errors.Is(err, ErrSlowDown):
		return identity.ErrSlowDown
	case errors.Is(err, ErrExpiredToken):
		return identity.ErrExpiredDeviceCode
	case errors.Is(err, ErrAccessDenied):
		return identity.ErrAccessDenied
	case errors.Is(err, ErrInvalidGrant):
		return identity.ErrSessionExpired
	default:
		return err
	}
}
