package identity

import (
	"context"
	"errors"

	"github.com/usetero/cli/internal/infrastructure/auth/workos"
)

// StartDeviceAuth starts device authorization through WorkOS.
func (p *WorkOSProvider) StartDeviceAuth(ctx context.Context) (DeviceFlow, error) {
	if p == nil || p.client == nil {
		return DeviceFlow{}, ErrProviderNotConfigured
	}
	flow, err := p.client.StartDeviceAuthorization(ctx)
	if err != nil {
		return DeviceFlow{}, mapWorkOSError(err)
	}
	return DeviceFlow{
		UserCode:                flow.UserCode,
		VerificationURI:         flow.VerificationURI,
		VerificationURIComplete: flow.VerificationURIComplete,
		DeviceCode:              flow.DeviceCode,
		ExpiresIn:               flow.ExpiresIn,
		Interval:                flow.Interval,
	}, nil
}

// PollAuthentication polls WorkOS for device auth completion.
func (p *WorkOSProvider) PollAuthentication(ctx context.Context, deviceCode string) (Tokens, User, error) {
	if p == nil || p.client == nil {
		return Tokens{}, User{}, ErrProviderNotConfigured
	}
	result, err := p.client.PollDeviceAuthorization(ctx, deviceCode)
	if err != nil {
		return Tokens{}, User{}, mapWorkOSError(err)
	}
	return Tokens{
			AccessToken:  AccessToken(result.AccessToken),
			RefreshToken: RefreshToken(result.RefreshToken),
		}, User{
			ID:            result.User.ID,
			Email:         result.User.Email,
			EmailVerified: result.User.EmailVerified,
			FirstName:     result.User.FirstName,
			LastName:      result.User.LastName,
		}, nil
}

// Refresh exchanges refresh token through WorkOS.
func (p *WorkOSProvider) Refresh(ctx context.Context, refreshToken RefreshToken, providerOrgID ProviderOrgID) (Tokens, error) {
	if p == nil || p.client == nil {
		return Tokens{}, ErrProviderNotConfigured
	}
	result, err := p.client.RefreshToken(ctx, string(refreshToken), string(providerOrgID))
	if err != nil {
		return Tokens{}, mapWorkOSError(err)
	}
	return Tokens{
		AccessToken:  AccessToken(result.AccessToken),
		RefreshToken: RefreshToken(result.RefreshToken),
	}, nil
}

func mapWorkOSError(err error) error {
	switch {
	case errors.Is(err, workos.ErrAuthorizationPending):
		return ErrAuthorizationPending
	case errors.Is(err, workos.ErrSlowDown):
		return ErrSlowDown
	case errors.Is(err, workos.ErrExpiredToken):
		return ErrExpiredDeviceCode
	case errors.Is(err, workos.ErrAccessDenied):
		return ErrAccessDenied
	default:
		return err
	}
}
