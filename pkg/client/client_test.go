package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestErrorCleaningClient(t *testing.T) {
	t.Run("extracts message from HTTPError", func(t *testing.T) {
		// When the control plane returns a non-200 status (e.g., 401),
		// genqlient wraps it in an HTTPError. The Error() method returns
		// raw JSON, but Response.Errors contains the parsed error messages.
		mockBase := &mockGraphQLClient{
			makeRequestFunc: func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
				return &graphql.HTTPError{
					Response: graphql.Response{
						Errors: gqlerror.List{
							{Message: "authentication required, see https://docs.usetero.com/api"},
						},
					},
				}
			},
		}

		client := &errorCleaningClient{base: mockBase}
		err := client.MakeRequest(context.Background(), &graphql.Request{}, &graphql.Response{})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		got := err.Error()
		want := "authentication required, see https://docs.usetero.com/api"

		if got != want {
			t.Errorf("error message = %q, want %q", got, want)
		}
	})

	t.Run("returns clean error messages without GraphQL path prefixes", func(t *testing.T) {
		// When the GraphQL API returns an error with a path (like "createDatadogAccount"),
		// gqlerror.Error() formats it as "input: createDatadogAccount <message>".
		// We extract just the message so users see clean, friendly errors.
		mockBase := &mockGraphQLClient{
			makeRequestFunc: func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
				return gqlerror.List{
					{
						Message: "Invalid Datadog credentials. Please verify your API key and Application key have the required permissions",
						Path:    ast.Path{ast.PathName("createDatadogAccount")},
					},
				}
			},
		}

		client := &errorCleaningClient{base: mockBase}
		err := client.MakeRequest(context.Background(), &graphql.Request{}, &graphql.Response{})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		got := err.Error()
		want := "Invalid Datadog credentials. Please verify your API key and Application key have the required permissions"

		if got != want {
			t.Errorf("error message = %q, want %q", got, want)
		}
	})
}

// mockGraphQLClient implements graphql.Client for testing
type mockGraphQLClient struct {
	makeRequestFunc func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error
}

func (m *mockGraphQLClient) MakeRequest(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
	if m.makeRequestFunc != nil {
		return m.makeRequestFunc(ctx, req, resp)
	}
	return nil
}

func TestAuthTransport_RefreshesOn401(t *testing.T) {
	t.Run("refreshes token and retries on 401", func(t *testing.T) {
		callCount := 0
		refreshCalled := false

		mockRT := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				callCount++
				if callCount == 1 {
					// First call: return 401
					return &http.Response{
						StatusCode: http.StatusUnauthorized,
						Body:       io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"token expired"}]}`)),
						Header:     make(http.Header),
					}, nil
				}
				// Second call: return success
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"data":{}}`)),
					Header:     make(http.Header),
				}, nil
			},
		}

		transport := &authTransport{
			accessToken: "expired-token",
			base:        mockRT,
			refreshFunc: func() (string, error) {
				refreshCalled = true
				return "new-token", nil
			},
		}

		req, _ := http.NewRequest("POST", "http://example.com/graphql", nil)
		resp, err := transport.RoundTrip(req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		if !refreshCalled {
			t.Error("expected refresh to be called")
		}
		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
		if transport.accessToken != "new-token" {
			t.Errorf("expected token to be updated to 'new-token', got %q", transport.accessToken)
		}
	})

	t.Run("does not refresh on non-401 errors", func(t *testing.T) {
		refreshCalled := false

		mockRT := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"server error"}]}`)),
					Header:     make(http.Header),
				}, nil
			},
		}

		transport := &authTransport{
			accessToken: "some-token",
			base:        mockRT,
			refreshFunc: func() (string, error) {
				refreshCalled = true
				return "new-token", nil
			},
		}

		req, _ := http.NewRequest("POST", "http://example.com/graphql", nil)
		resp, _ := transport.RoundTrip(req)

		if refreshCalled {
			t.Error("refresh should not be called for non-401 errors")
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	t.Run("works without refresh func (backwards compatible)", func(t *testing.T) {
		mockRT := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"token expired"}]}`)),
					Header:     make(http.Header),
				}, nil
			},
		}

		transport := &authTransport{
			accessToken: "expired-token",
			base:        mockRT,
			// No refreshFunc - should just return the 401
		}

		req, _ := http.NewRequest("POST", "http://example.com/graphql", nil)
		resp, err := transport.RoundTrip(req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})
}

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return nil, nil
}
