package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	genqlient "github.com/Khan/genqlient/graphql"
)

const defaultTimeout = 30 * time.Second
const graphqlPath = "/graphql"

type baseClient struct {
	origin string
	http   *http.Client
	token  TokenProvider
}

// BootstrapClient exposes control-plane operations that are available before
// an account has been selected.
type BootstrapClient struct {
	base *baseClient
}

// AccountClient exposes account-owned control-plane operations and always
// sends the account scope header.
type AccountClient struct {
	base      *baseClient
	accountID AccountID
}

// NewBootstrapClient creates a new unscoped control-plane API client.
func NewBootstrapClient(origin string, token TokenProvider) *BootstrapClient {
	origin, err := normalizeOrigin(origin)
	if err != nil {
		panic(fmt.Sprintf("controlplane api client requires valid origin: %v", err))
	}
	if origin == "" {
		panic("controlplane api client requires origin")
	}
	return &BootstrapClient{
		base: &baseClient{
			origin: origin,
			http:   &http.Client{Timeout: defaultTimeout},
			token:  token,
		},
	}
}

// ForAccount binds account scope once so subsequent operations cannot forget it.
func (c *BootstrapClient) ForAccount(accountID AccountID) (*AccountClient, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}
	return &AccountClient{base: c.base, accountID: accountID}, nil
}

func (c *baseClient) gql(ctx context.Context) (genqlient.Client, error) {
	return c.gqlScoped(ctx, "")
}

func (c *baseClient) gqlForAccount(ctx context.Context, accountID AccountID) (genqlient.Client, error) {
	return c.gqlScoped(ctx, accountID.String())
}

func (c *baseClient) gqlScoped(ctx context.Context, accountID string) (genqlient.Client, error) {
	token := ""
	if c.token != nil {
		t, err := c.token.GetAccessToken(ctx)
		if err != nil {
			return nil, err
		}
		token = t
	}

	transport := c.http.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	httpClient := &http.Client{
		Timeout: c.http.Timeout,
		Transport: &authTransport{
			token:     token,
			accountID: accountID,
			base:      transport,
		},
	}

	return genqlient.NewClient(c.graphQLEndpoint(), httpClient), nil
}

func (c *baseClient) graphQLEndpoint() string {
	return c.origin + graphqlPath
}

func normalizeOrigin(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("origin is required")
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid origin %q", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("origin must use http or https: %q", raw)
	}
	if u.Path != "" && u.Path != "/" {
		return "", fmt.Errorf("origin must not include a path: %q", raw)
	}
	if u.RawQuery != "" {
		return "", fmt.Errorf("origin must not include a query: %q", raw)
	}
	if u.Fragment != "" {
		return "", fmt.Errorf("origin must not include a fragment: %q", raw)
	}

	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

type authTransport struct {
	token     string
	accountID string
	base      http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	if t.accountID != "" {
		req.Header.Set("X-Account-ID", t.accountID)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
