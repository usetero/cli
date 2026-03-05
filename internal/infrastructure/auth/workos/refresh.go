package workos

import (
	"context"
	"fmt"
	"net/url"
)

// RefreshToken exchanges a refresh token for a new token pair.
// If workosOrgID is non-empty, the token is scoped to that organization.
func (c *Client) RefreshToken(ctx context.Context, refreshToken, workosOrgID string) (RefreshResult, error) {
	endpoint := fmt.Sprintf("%s/user_management/authenticate", c.baseURL)
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")
	if workosOrgID != "" {
		form.Set("organization_id", workosOrgID)
	}
	for _, aud := range c.audiences {
		form.Add("audience", aud)
	}

	var out authenticateResponse
	if err := c.doForm(ctx, endpoint, form, &out); err != nil {
		return RefreshResult{}, err
	}
	return RefreshResult{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}, nil
}
