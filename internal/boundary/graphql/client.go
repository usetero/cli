package graphql

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
)

const (
	retryMax          = 3
	retryWaitMin      = 100 * time.Millisecond
	retryWaitMax      = 2 * time.Second
	defaultAPITimeout = 30 * time.Second
)

// Client defines the interface for communicating with the Tero control plane.
// This allows services to be tested without real API calls.
type Client interface {
	// SetAccountID sets the account ID header for scoped requests.
	SetAccountID(accountID domain.AccountID)
	// WithAccountID returns a new client scoped to accountID.
	WithAccountID(accountID domain.AccountID) Client

	// RawQuery executes an arbitrary GraphQL query (for debugging).
	RawQuery(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error)

	// Organization operations
	ListOrganizations(ctx context.Context) (*gen.ListOrganizationsResponse, error)
	CreateOrganizationAndBootstrap(ctx context.Context, input gen.OrganizationCreateInput) (*gen.CreateOrganizationAndBootstrapResponse, error)

	// Account operations
	ListAccounts(ctx context.Context, organizationID string) (*gen.ListAccountsResponse, error)
	CreateAccount(ctx context.Context, input gen.AccountCreateInput) (*gen.CreateAccountResponse, error)
	GetAccount(ctx context.Context, accountID string) (*gen.GetAccountResponse, error)

	// Datadog operations
	ValidateDatadogApiKey(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error)
	CreateDatadogAccountWithCredentials(ctx context.Context, input gen.DatadogAccountCreateInput) (*gen.CreateDatadogAccountWithCredentialsResponse, error)
	GetDatadogAccountStatus(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error)

	// Service operations
	EnableService(ctx context.Context, serviceID string) (*gen.EnableServiceResponse, error)
	DisableService(ctx context.Context, serviceID string) (*gen.DisableServiceResponse, error)

	// Product surface reads
	GetIssueSummary(ctx context.Context) (*gen.GetIssueSummaryResponse, error)
	ListChecks(ctx context.Context) (*gen.ListChecksResponse, error)
	ListEdgeInstances(ctx context.Context) (*gen.ListEdgeInstancesResponse, error)

	// Data-plane status reads
	GetAccountStatusSummary(ctx context.Context) (*gen.GetAccountStatusSummaryResponse, error)
	ListServiceStatuses(ctx context.Context, first int) (*gen.ListServiceStatusesResponse, error)
	ListServiceLogEvents(ctx context.Context, serviceID string, first int) (*gen.ListServiceLogEventsResponse, error)
}

// client is the concrete implementation of Client.
type client struct {
	endpoint  string
	auth      auth.Auth
	http      *http.Client
	accountID string
	mu        sync.RWMutex
}

// Ensure client implements Client.
var _ Client = (*client)(nil)

// NewClient creates a new GraphQL API client.
// - Retries transient errors (connection reset, 502/503/504) up to 3 times with backoff
// - Gets a fresh token via auth.GetAccessToken before each request
func NewClient(endpoint string, authService auth.Auth) Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Logger = nil

	httpClient := retryClient.StandardClient()
	httpClient.Timeout = defaultAPITimeout

	return &client{
		endpoint: endpoint,
		auth:     authService,
		http:     httpClient,
	}
}

// SetAccountID sets the account ID header for scoped requests.
func (c *client) SetAccountID(accountID domain.AccountID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accountID = accountID.String()
}

// WithAccountID returns a copy scoped to the given account.
func (c *client) WithAccountID(accountID domain.AccountID) Client {
	return &client{
		endpoint:  c.endpoint,
		auth:      c.auth,
		http:      c.http,
		accountID: accountID.String(),
	}
}

// gql returns a graphql.Client configured with fresh auth token.
func (c *client) gql(ctx context.Context) (graphql.Client, error) {
	token, err := c.auth.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	accountID := c.accountIDSnapshot()
	timeout := c.http.Timeout
	if timeout <= 0 {
		timeout = defaultAPITimeout
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &authTransport{
			token:     token,
			accountID: accountID,
			base:      c.http.Transport,
		},
	}

	return graphql.NewClient(c.endpoint, httpClient), nil
}

// authTransport adds Authorization and Account ID headers to requests.
type authTransport struct {
	token     string
	accountID string
	base      http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	if t.accountID != "" {
		req.Header.Set("X-Account-ID", t.accountID)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

func (c *client) accountIDSnapshot() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accountID
}

// RawQuery executes an arbitrary GraphQL query and returns the raw result.
func (c *client) RawQuery(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}

	req := &graphql.Request{
		Query:     query,
		Variables: variables,
	}
	resp := &graphql.Response{
		Data: new(map[string]interface{}),
	}

	if err := gql.MakeRequest(ctx, req, resp); err != nil {
		return nil, err
	}

	if data, ok := resp.Data.(*map[string]interface{}); ok {
		return *data, nil
	}
	return nil, nil
}

// Organization operations

func (c *client) ListOrganizations(ctx context.Context) (*gen.ListOrganizationsResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListOrganizations(ctx, gql)
}

func (c *client) CreateOrganizationAndBootstrap(ctx context.Context, input gen.OrganizationCreateInput) (*gen.CreateOrganizationAndBootstrapResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.CreateOrganizationAndBootstrap(ctx, gql, input)
}

// Account operations

func (c *client) ListAccounts(ctx context.Context, organizationID string) (*gen.ListAccountsResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListAccounts(ctx, gql, organizationID)
}

func (c *client) CreateAccount(ctx context.Context, input gen.AccountCreateInput) (*gen.CreateAccountResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.CreateAccount(ctx, gql, input)
}

func (c *client) GetAccount(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.GetAccount(ctx, gql, accountID)
}

// Datadog operations

//nolint:staticcheck // matches generated GraphQL function name
func (c *client) ValidateDatadogApiKey(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ValidateDatadogApiKey(ctx, gql, input)
}

func (c *client) CreateDatadogAccountWithCredentials(ctx context.Context, input gen.DatadogAccountCreateInput) (*gen.CreateDatadogAccountWithCredentialsResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.CreateDatadogAccountWithCredentials(ctx, gql, input)
}

func (c *client) GetDatadogAccountStatus(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.GetDatadogAccountStatus(ctx, gql, id)
}

// Service operations

func (c *client) EnableService(ctx context.Context, serviceID string) (*gen.EnableServiceResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.EnableService(ctx, gql, serviceID)
}

func (c *client) DisableService(ctx context.Context, serviceID string) (*gen.DisableServiceResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.DisableService(ctx, gql, serviceID)
}

// Product surface reads

func (c *client) GetIssueSummary(ctx context.Context) (*gen.GetIssueSummaryResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.GetIssueSummary(ctx, gql)
}

func (c *client) ListChecks(ctx context.Context) (*gen.ListChecksResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListChecks(ctx, gql)
}

func (c *client) ListEdgeInstances(ctx context.Context) (*gen.ListEdgeInstancesResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListEdgeInstances(ctx, gql)
}

// Data-plane status reads

func (c *client) GetAccountStatusSummary(ctx context.Context) (*gen.GetAccountStatusSummaryResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.GetAccountStatusSummary(ctx, gql)
}

func (c *client) ListServiceStatuses(ctx context.Context, first int) (*gen.ListServiceStatusesResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListServiceStatuses(ctx, gql, first)
}

func (c *client) ListServiceLogEvents(ctx context.Context, serviceID string, first int) (*gen.ListServiceLogEventsResponse, error) {
	gql, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListServiceLogEvents(ctx, gql, serviceID, first)
}
