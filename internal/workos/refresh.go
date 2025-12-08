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

// RefreshToken exchanges a refresh token for a new access token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshResponse, error) {
	return c.refreshToken(ctx, refreshToken, "")
}

// RefreshTokenWithOrganization exchanges a refresh token for a new access token scoped to an organization.
// The organizationID should be the WorkOS organization ID.
func (c *Client) RefreshTokenWithOrganization(ctx context.Context, refreshToken, organizationID string) (*auth.RefreshResponse, error) {
	return c.refreshToken(ctx, refreshToken, organizationID)
}

func (c *Client) refreshToken(ctx context.Context, refreshToken, organizationID string) (*auth.RefreshResponse, error) {
	endpoint := fmt.Sprintf("%s/user_management/authenticate", c.baseURL)

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	if organizationID != "" {
		data.Set("organization_id", organizationID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &auth.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
