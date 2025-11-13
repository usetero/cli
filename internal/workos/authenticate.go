package workos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AuthResponse contains the response from successful authentication.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	User         User   `json:"user"`
}

// pollAuthentication polls WorkOS to check if the user has completed authentication.
// Returns AuthResponse on success, or a specific error type based on the response.
// This is the internal implementation - use PollAuthentication from adapter.go for app.AuthClient interface.
func (c *Client) pollAuthentication(ctx context.Context, deviceCode string) (*AuthResponse, error) {
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
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, parseError(errResp.Error, errResp.ErrorDescription, resp.StatusCode)
		}
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result AuthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// parseError converts WorkOS error codes into specific error types.
func parseError(code, description string, statusCode int) error {
	switch code {
	case "authorization_pending":
		return &AuthorizationPendingError{}
	case "slow_down":
		return &SlowDownError{}
	case "expired_token":
		return &ExpiredTokenError{}
	case "access_denied":
		return &AccessDeniedError{}
	default:
		return &UnknownError{
			Code:        code,
			Description: description,
			StatusCode:  statusCode,
		}
	}
}
