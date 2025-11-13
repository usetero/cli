package client

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Client struct {
	gql graphql.Client
}

// New creates a new authenticated GraphQL client.
// The accessToken is added to all requests via Authorization header.
func New(endpoint string, accessToken string) *Client {
	httpClient := &http.Client{
		Transport: &authTransport{
			accessToken: accessToken,
			base:        http.DefaultTransport,
		},
	}

	baseClient := graphql.NewClient(endpoint, httpClient)

	return &Client{
		gql: &errorCleaningClient{base: baseClient},
	}
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

// authTransport adds Authorization header to all requests
type authTransport struct {
	accessToken string
	base        http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original
	req = req.Clone(req.Context())

	// Add Authorization header
	req.Header.Set("Authorization", "Bearer "+t.accessToken)

	// Execute request
	return t.base.RoundTrip(req)
}
