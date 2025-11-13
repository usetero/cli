package client

import (
	"context"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestErrorCleaningClient(t *testing.T) {
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
