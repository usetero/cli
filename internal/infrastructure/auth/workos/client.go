package workos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.workos.com"
	defaultTimeout = 30 * time.Second
)

// Client talks to WorkOS OAuth device flow endpoints.
type Client struct {
	baseURL   string
	clientID  string
	audiences []string
	http      *http.Client
}

// Option mutates WorkOS client construction behavior.
type Option func(*Client)

// WithBaseURL overrides the default WorkOS API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.http = httpClient
		}
	}
}

// NewClient creates a WorkOS client.
func NewClient(clientID string, audiences []string, opts ...Option) (*Client, error) {
	if clientID == "" {
		return nil, fmt.Errorf("workos client id is required")
	}
	if len(audiences) == 0 {
		return nil, fmt.Errorf("at least one audience is required")
	}

	c := &Client{
		baseURL:   defaultBaseURL,
		clientID:  clientID,
		audiences: append([]string(nil), audiences...),
		http:      &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) doForm(ctx context.Context, endpoint string, values url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var oauthErr oauthErrorResponse
		if err := json.Unmarshal(body, &oauthErr); err == nil && oauthErr.Error != "" {
			return &OAuthError{Code: oauthErr.Error, Description: oauthErr.ErrorDescription}
		}
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if err := json.Unmarshal(body, out); err != nil {
		return err
	}
	return nil
}
