package workos

import (
	"context"
	"errors"

	"github.com/usetero/cli/internal/auth"
)

// AuthorizeDevice implements auth.OAuthProvider interface.
// Converts WorkOS response to auth.DeviceAuthResponse.
func (c *Client) AuthorizeDevice(ctx context.Context) (*auth.DeviceAuthResponse, error) {
	resp, err := c.authorizeDevice(ctx)
	if err != nil {
		return nil, err
	}

	return &auth.DeviceAuthResponse{
		DeviceCode:              resp.DeviceCode,
		UserCode:                resp.UserCode,
		VerificationURI:         resp.VerificationURI,
		VerificationURIComplete: resp.VerificationURIComplete,
		ExpiresIn:               resp.ExpiresIn,
		Interval:                resp.Interval,
	}, nil
}

// PollAuthentication implements auth.OAuthProvider interface.
// Converts WorkOS response to auth.AuthenticationResponse.
func (c *Client) PollAuthentication(ctx context.Context, deviceCode string) (*auth.AuthenticationResponse, error) {
	resp, err := c.pollAuthentication(ctx, deviceCode)
	if err != nil {
		// Convert WorkOS errors to auth errors
		return nil, convertError(err)
	}

	return &auth.AuthenticationResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User: auth.User{
			ID:            resp.User.ID,
			Email:         resp.User.Email,
			EmailVerified: resp.User.EmailVerified,
			FirstName:     resp.User.FirstName,
			LastName:      resp.User.LastName,
		},
	}, nil
}

// convertError converts WorkOS errors to auth errors.
func convertError(err error) error {
	var pendingErr *AuthorizationPendingError
	var slowDownErr *SlowDownError
	var expiredErr *ExpiredTokenError
	var deniedErr *AccessDeniedError

	if errors.As(err, &pendingErr) {
		return &auth.AuthorizationPendingError{}
	}
	if errors.As(err, &slowDownErr) {
		return &auth.SlowDownError{}
	}
	if errors.As(err, &expiredErr) {
		return &auth.ExpiredTokenError{}
	}
	if errors.As(err, &deniedErr) {
		return &auth.AccessDeniedError{}
	}
	return err
}
