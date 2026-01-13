package client

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Client struct {
	gql       graphql.Client
	transport *authTransport
}

// New creates a new authenticated GraphQL client with automatic token refresh and retry.
// - Retries transient errors (connection reset, 502/503/504) up to 3 times with backoff
// - When a request returns 401, the refreshFunc is called to get a new token and the request is retried
func New(endpoint string, accessToken string, refreshFunc RefreshFunc) *Client {
	// Configure retryable HTTP client
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 100 * time.Millisecond
	retryClient.RetryWaitMax = 2 * time.Second
	retryClient.Logger = nil // Disable logging

	// Wrap with auth transport for token handling
	transport := &authTransport{
		accessToken: accessToken,
		base:        retryClient.StandardClient().Transport,
		refreshFunc: refreshFunc,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	baseClient := graphql.NewClient(endpoint, httpClient)

	return &Client{
		gql:       &errorCleaningClient{base: baseClient},
		transport: transport,
	}
}

// SetAccessToken updates the access token used for authentication.
func (c *Client) SetAccessToken(token string) {
	c.transport.SetAccessToken(token)
}

// errorCleaningClient wraps a graphql.Client and cleans up error messages
// by removing GraphQL-specific prefixes like "input: operationName".
type errorCleaningClient struct {
	base graphql.Client
}

// MakeRequest implements graphql.Client by delegating to the base client
// and cleaning any errors that are returned.
func (c *errorCleaningClient) MakeRequest(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
	err := c.base.MakeRequest(ctx, req, resp)
	if err != nil {
		return cleanGraphQLError(err)
	}
	return nil
}

// cleanGraphQLError removes GraphQL-specific prefixes from error messages.
// gqlerror.Error.Error() formats errors as "input: <path> <message>".
// We strip the "input: <path>" prefix to show clean user-friendly messages.
func cleanGraphQLError(err error) error {
	if err == nil {
		return nil
	}

	// Handle HTTPError (non-200 status codes)
	// genqlient wraps HTTP errors with the raw JSON response in Error(),
	// but the parsed errors are available in Response.Errors
	var httpErr *graphql.HTTPError
	if errors.As(err, &httpErr) {
		if len(httpErr.Response.Errors) > 0 {
			return errors.New(httpErr.Response.Errors[0].Message)
		}
	}

	// Handle gqlerror.List (multiple errors)
	var gqlErrList gqlerror.List
	if errors.As(err, &gqlErrList) {
		cleaned := make([]string, 0, len(gqlErrList))
		for _, gqlErr := range gqlErrList {
			// Use the Message field directly instead of Error() which adds prefixes
			if gqlErr.Message != "" {
				cleaned = append(cleaned, gqlErr.Message)
			}
		}
		if len(cleaned) > 0 {
			return errors.New(strings.Join(cleaned, "\n"))
		}
	}

	// Handle single gqlerror.Error
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) && gqlErr.Message != "" {
		return errors.New(gqlErr.Message)
	}

	// Fallback to original error
	return err
}

// RefreshFunc is a function that refreshes the access token.
// It returns the new token or an error.
type RefreshFunc func() (string, error)

// authTransport adds Authorization header to all requests
// and automatically refreshes the token on 401 responses.
type authTransport struct {
	accessToken string
	base        http.RoundTripper
	refreshFunc RefreshFunc
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original
	req = req.Clone(req.Context())

	// Add Authorization header
	req.Header.Set("Authorization", "Bearer "+t.accessToken)

	// Execute request
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// If 401 and we have a refresh function, try to refresh and retry
	if resp.StatusCode == http.StatusUnauthorized && t.refreshFunc != nil {
		// Close the original response body
		resp.Body.Close()

		// Try to refresh the token
		newToken, refreshErr := t.refreshFunc()
		if refreshErr != nil {
			// Refresh failed, return the original 401
			// Re-execute to get a fresh response since we closed the body
			return t.base.RoundTrip(req)
		}

		// Update token and retry
		t.accessToken = newToken
		req.Header.Set("Authorization", "Bearer "+t.accessToken)
		return t.base.RoundTrip(req)
	}

	return resp, nil
}

// SetAccessToken updates the access token used for authentication.
func (t *authTransport) SetAccessToken(token string) {
	t.accessToken = token
}
