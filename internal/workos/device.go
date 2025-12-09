package workos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/usetero/cli/internal/auth"
)

// AuthorizeDevice initiates the device authorization flow.
// Returns device_code (for polling), user_code (to show user), and verification URLs.
func (c *Client) AuthorizeDevice(ctx context.Context) (*auth.DeviceAuthResponse, error) {
	endpoint := fmt.Sprintf("%s/user_management/authorize/device", c.baseURL)

	data := url.Values{}
	data.Set("client_id", c.clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	// WorkOS response format
	var result struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &auth.DeviceAuthResponse{
		DeviceCode:              result.DeviceCode,
		UserCode:                result.UserCode,
		VerificationURI:         result.VerificationURI,
		VerificationURIComplete: result.VerificationURIComplete,
		ExpiresIn:               result.ExpiresIn,
		Interval:                result.Interval,
	}, nil
}

// PollAuthentication polls WorkOS to check if the user has completed authentication.
func (c *Client) PollAuthentication(ctx context.Context, deviceCode string) (*auth.AuthenticationResponse, error) {
	endpoint := fmt.Sprintf("%s/user_management/authenticate", c.baseURL)

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, parseError(errResp.Error)
		}
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	// WorkOS response format
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID            string `json:"id"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			FirstName     string `json:"first_name"`
			LastName      string `json:"last_name"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &auth.AuthenticationResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: auth.User{
			ID:            result.User.ID,
			Email:         result.User.Email,
			EmailVerified: result.User.EmailVerified,
			FirstName:     result.User.FirstName,
			LastName:      result.User.LastName,
		},
	}, nil
}

// parseError converts WorkOS error codes into auth error types.
func parseError(code string) error {
	switch code {
	case "authorization_pending":
		return &auth.AuthorizationPendingError{}
	case "slow_down":
		return &auth.SlowDownError{}
	case "expired_token":
		return &auth.ExpiredTokenError{}
	case "access_denied":
		return &auth.AccessDeniedError{}
	default:
		return fmt.Errorf("workos error: %s", code)
	}
}
