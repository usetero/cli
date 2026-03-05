package workos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// StartDeviceAuthorization initiates WorkOS device authorization flow.
func (c *Client) StartDeviceAuthorization(ctx context.Context) (DeviceAuthorization, error) {
	endpoint := fmt.Sprintf("%s/user_management/authorize/device", c.baseURL)
	form := url.Values{}
	form.Set("client_id", c.clientID)

	var out deviceAuthResponse
	if err := c.doForm(ctx, endpoint, form, &out); err != nil {
		return DeviceAuthorization{}, err
	}

	return DeviceAuthorization{
		UserCode:                out.UserCode,
		VerificationURI:         out.VerificationURI,
		VerificationURIComplete: out.VerificationURIComplete,
		DeviceCode:              out.DeviceCode,
		ExpiresIn:               time.Duration(out.ExpiresIn) * time.Second,
		Interval:                time.Duration(out.Interval) * time.Second,
	}, nil
}

// PollDeviceAuthorization polls WorkOS until authentication completes.
func (c *Client) PollDeviceAuthorization(ctx context.Context, deviceCode string) (DeviceAuthenticationResult, error) {
	endpoint := fmt.Sprintf("%s/user_management/authenticate", c.baseURL)
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("device_code", deviceCode)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	for _, aud := range c.audiences {
		form.Add("audience", aud)
	}

	var out authenticateResponse
	if err := c.doForm(ctx, endpoint, form, &out); err != nil {
		return DeviceAuthenticationResult{}, err
	}

	return DeviceAuthenticationResult{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		User: AuthenticatedUser{
			ID:            out.User.ID,
			Email:         out.User.Email,
			EmailVerified: out.User.EmailVerified,
			FirstName:     out.User.FirstName,
			LastName:      out.User.LastName,
		},
	}, nil
}
