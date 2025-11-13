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

// DeviceAuthResponse contains the response from the device authorization request.
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// authorizeDevice initiates the device authorization flow.
// Returns device_code (for polling), user_code (to show user), and verification URLs.
// This is the internal implementation - use AuthorizeDevice from adapter.go for app.AuthClient interface.
func (c *Client) authorizeDevice(ctx context.Context) (*DeviceAuthResponse, error) {
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
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result DeviceAuthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
