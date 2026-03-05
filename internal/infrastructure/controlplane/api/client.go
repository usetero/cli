package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	genqlient "github.com/Khan/genqlient/graphql"
	"github.com/usetero/cli/internal/infrastructure/controlplane/api/gen"
)

const defaultTimeout = 30 * time.Second

// TokenProvider returns a control-plane bearer token for API requests.
type TokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

type OrganizationID string

func (id OrganizationID) String() string { return string(id) }

type AccountID string

func (id AccountID) String() string { return string(id) }

type WorkspaceID string

func (id WorkspaceID) String() string { return string(id) }

type DatadogAccountID string

func (id DatadogAccountID) String() string { return string(id) }

// Organization is the control-plane organization payload.
type Organization struct {
	ID                   OrganizationID
	Name                 string
	WorkosOrganizationID string
}

// Account is the control-plane account payload.
type Account struct {
	ID        AccountID
	Name      string
	CreatedAt time.Time
}

// Workspace is the control-plane workspace payload.
type Workspace struct {
	ID        WorkspaceID
	Name      string
	CreatedAt time.Time
}

// OrganizationBootstrap is the create-organization bootstrap payload.
type OrganizationBootstrap struct {
	Organization Organization
	Account      Account
	Workspace    Workspace
}

// DatadogSite is the Datadog regional site enum.
type DatadogSite string

const (
	DatadogSiteUS1    DatadogSite = DatadogSite(gen.DatadogAccountSiteUs1)
	DatadogSiteUS3    DatadogSite = DatadogSite(gen.DatadogAccountSiteUs3)
	DatadogSiteUS5    DatadogSite = DatadogSite(gen.DatadogAccountSiteUs5)
	DatadogSiteEU1    DatadogSite = DatadogSite(gen.DatadogAccountSiteEu1)
	DatadogSiteUS1Fed DatadogSite = DatadogSite(gen.DatadogAccountSiteUs1Fed)
	DatadogSiteAP1    DatadogSite = DatadogSite(gen.DatadogAccountSiteAp1)
	DatadogSiteAP2    DatadogSite = DatadogSite(gen.DatadogAccountSiteAp2)
)

func (s DatadogSite) Valid() bool {
	switch s {
	case DatadogSiteUS1, DatadogSiteUS3, DatadogSiteUS5, DatadogSiteEU1, DatadogSiteUS1Fed, DatadogSiteAP1, DatadogSiteAP2:
		return true
	default:
		return false
	}
}

type DatadogAccount struct {
	ID   DatadogAccountID
	Name string
	Site DatadogSite
}

type DatadogAccountHealth string

const (
	DatadogAccountHealthDisabled DatadogAccountHealth = DatadogAccountHealth(gen.DatadogAccountStatusCacheHealthDisabled)
	DatadogAccountHealthInactive DatadogAccountHealth = DatadogAccountHealth(gen.DatadogAccountStatusCacheHealthInactive)
	DatadogAccountHealthError    DatadogAccountHealth = DatadogAccountHealth(gen.DatadogAccountStatusCacheHealthError)
	DatadogAccountHealthOK       DatadogAccountHealth = DatadogAccountHealth(gen.DatadogAccountStatusCacheHealthOk)
)

type DatadogAccountStatus struct {
	Health               DatadogAccountHealth
	ReadyForUse          bool
	ServiceCount         int
	ActiveServices       int
	OKServices           int
	DisabledServices     int
	InactiveServices     int
	EventCount           int
	AnalyzedCount        int
	PendingPolicyCount   int
	ApprovedPolicyCount  int
	DismissedPolicyCount int
}

type CreateDatadogAccountInput struct {
	AccountID AccountID
	Name      string
	Site      DatadogSite
	APIKey    string
	AppKey    string
}

// Client wraps typed genqlient operations for control-plane APIs.
type Client struct {
	endpoint string
	http     *http.Client
	token    TokenProvider
}

// NewClient creates a new control-plane API client.
func NewClient(endpoint string, token TokenProvider) *Client {
	return &Client{
		endpoint: endpoint,
		http:     &http.Client{Timeout: defaultTimeout},
		token:    token,
	}
}

// ListOrganizations fetches organizations for the current user.
func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	if c == nil {
		return nil, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return nil, fmt.Errorf("api endpoint is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.ListOrganizations(ctx, client)
	if err != nil {
		return nil, err
	}

	edges := resp.Organizations.Edges
	out := make([]Organization, 0, len(edges))
	for _, edge := range edges {
		if edge == nil {
			continue
		}
		node := edge.Node
		out = append(out, Organization{
			ID:                   OrganizationID(node.Id),
			Name:                 node.Name,
			WorkosOrganizationID: node.WorkosOrganizationID,
		})
	}
	return out, nil
}

// ListAccounts fetches accounts for the given organization.
func (c *Client) ListAccounts(ctx context.Context, organizationID OrganizationID) ([]Account, error) {
	if c == nil {
		return nil, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return nil, fmt.Errorf("api endpoint is required")
	}
	if organizationID == "" {
		return nil, fmt.Errorf("organization id is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.ListAccounts(ctx, client, organizationID.String())
	if err != nil {
		return nil, err
	}

	edges := resp.Accounts.Edges
	out := make([]Account, 0, len(edges))
	for _, edge := range edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		node := edge.Node
		out = append(out, Account{
			ID:        AccountID(node.Id),
			Name:      node.Name,
			CreatedAt: node.CreatedAt,
		})
	}
	return out, nil
}

// CreateAccount creates an account in the given organization.
func (c *Client) CreateAccount(ctx context.Context, organizationID OrganizationID, name string) (Account, error) {
	if c == nil {
		return Account{}, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return Account{}, fmt.Errorf("api endpoint is required")
	}
	if organizationID == "" {
		return Account{}, fmt.Errorf("organization id is required")
	}
	if name == "" {
		return Account{}, fmt.Errorf("account name is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return Account{}, err
	}

	resp, err := gen.CreateAccount(ctx, client, gen.CreateAccountInput{
		Name:           name,
		OrganizationID: organizationID.String(),
	})
	if err != nil {
		return Account{}, err
	}

	return Account{
		ID:        AccountID(resp.CreateAccount.Id),
		Name:      resp.CreateAccount.Name,
		CreatedAt: resp.CreateAccount.CreatedAt,
	}, nil
}

// ListWorkspaces fetches workspaces for the given account.
func (c *Client) ListWorkspaces(ctx context.Context, accountID AccountID) ([]Workspace, error) {
	if c == nil {
		return nil, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return nil, fmt.Errorf("api endpoint is required")
	}
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.ListWorkspaces(ctx, client, accountID.String())
	if err != nil {
		return nil, err
	}

	edges := resp.Workspaces.Edges
	out := make([]Workspace, 0, len(edges))
	for _, edge := range edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		node := edge.Node
		out = append(out, Workspace{
			ID:        WorkspaceID(node.Id),
			Name:      node.Name,
			CreatedAt: node.CreatedAt,
		})
	}
	return out, nil
}

// CreateOrganizationAndBootstrap creates org/account/workspace in one mutation.
func (c *Client) CreateOrganizationAndBootstrap(ctx context.Context, name string) (OrganizationBootstrap, error) {
	if c == nil {
		return OrganizationBootstrap{}, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return OrganizationBootstrap{}, fmt.Errorf("api endpoint is required")
	}
	if name == "" {
		return OrganizationBootstrap{}, fmt.Errorf("organization name is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	resp, err := gen.CreateOrganizationAndBootstrap(ctx, client, gen.CreateOrganizationInput{Name: name})
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	result := resp.CreateOrganizationAndBootstrap
	return OrganizationBootstrap{
		Organization: Organization{
			ID:                   OrganizationID(result.Organization.Id),
			Name:                 result.Organization.Name,
			WorkosOrganizationID: result.Organization.WorkosOrganizationID,
		},
		Account: Account{
			ID:   AccountID(result.Account.Id),
			Name: result.Account.Name,
		},
		Workspace: Workspace{
			ID:   WorkspaceID(result.Workspace.Id),
			Name: result.Workspace.Name,
		},
	}, nil
}

// GetAccountDatadogAccount fetches Datadog integration metadata for an account.
func (c *Client) GetAccountDatadogAccount(ctx context.Context, accountID AccountID) (*DatadogAccount, error) {
	if c == nil {
		return nil, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return nil, fmt.Errorf("api endpoint is required")
	}
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.GetAccount(ctx, client, accountID.String())
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.Accounts.Edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		if edge.Node.DatadogAccount == nil {
			return nil, nil
		}
		dd := edge.Node.DatadogAccount
		return &DatadogAccount{ID: DatadogAccountID(dd.Id), Name: dd.Name, Site: DatadogSite(dd.Site)}, nil
	}

	return nil, nil
}

// ValidateDatadogAPIKey asks control plane to validate a Datadog API key.
func (c *Client) ValidateDatadogAPIKey(ctx context.Context, apiKey string, site DatadogSite) (bool, string, error) {
	if c == nil {
		return false, "", fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return false, "", fmt.Errorf("api endpoint is required")
	}
	if apiKey == "" {
		return false, "", fmt.Errorf("api key is required")
	}
	if !site.Valid() {
		return false, "", fmt.Errorf("datadog site is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return false, "", err
	}

	resp, err := gen.ValidateDatadogApiKey(ctx, client, gen.ValidateDatadogApiKeyInput{
		ApiKey: apiKey,
		Site:   gen.DatadogAccountSite(site),
	})
	if err != nil {
		return false, "", err
	}

	out := resp.ValidateDatadogApiKey
	if out.Valid {
		return true, "", nil
	}

	message := "invalid api key"
	if out.Error != nil && *out.Error != "" {
		message = *out.Error
	}
	return false, message, nil
}

// CreateDatadogAccountWithCredentials creates Datadog integration with credentials.
func (c *Client) CreateDatadogAccountWithCredentials(ctx context.Context, input CreateDatadogAccountInput) (DatadogAccount, error) {
	if c == nil {
		return DatadogAccount{}, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return DatadogAccount{}, fmt.Errorf("api endpoint is required")
	}
	if input.AccountID == "" {
		return DatadogAccount{}, fmt.Errorf("account id is required")
	}
	if input.Name == "" {
		return DatadogAccount{}, fmt.Errorf("name is required")
	}
	if !input.Site.Valid() {
		return DatadogAccount{}, fmt.Errorf("datadog site is required")
	}
	if input.APIKey == "" {
		return DatadogAccount{}, fmt.Errorf("api key is required")
	}
	if input.AppKey == "" {
		return DatadogAccount{}, fmt.Errorf("app key is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return DatadogAccount{}, err
	}

	resp, err := gen.CreateDatadogAccountWithCredentials(ctx, client, gen.CreateDatadogAccountWithCredentialsInput{
		Attributes: gen.CreateDatadogAccountInput{
			AccountID: input.AccountID.String(),
			Name:      input.Name,
			Site:      gen.DatadogAccountSite(input.Site),
		},
		Credentials: gen.CreateDatadogCredentialsInput{
			ApiKey: input.APIKey,
			AppKey: input.AppKey,
		},
	})
	if err != nil {
		return DatadogAccount{}, err
	}

	created := resp.CreateDatadogAccount
	return DatadogAccount{ID: DatadogAccountID(created.Id), Name: created.Name, Site: DatadogSite(created.Site)}, nil
}

// GetDatadogAccountStatus returns status cache for Datadog account, or nil when unavailable.
func (c *Client) GetDatadogAccountStatus(ctx context.Context, datadogAccountID DatadogAccountID) (*DatadogAccountStatus, error) {
	if c == nil {
		return nil, fmt.Errorf("api client is nil")
	}
	if c.endpoint == "" {
		return nil, fmt.Errorf("api endpoint is required")
	}
	if datadogAccountID == "" {
		return nil, fmt.Errorf("datadog account id is required")
	}

	client, err := c.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.GetDatadogAccountStatus(ctx, client, datadogAccountID.String())
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.DatadogAccounts.Edges {
		if edge == nil || edge.Node == nil || edge.Node.Status == nil {
			continue
		}
		s := edge.Node.Status
		return &DatadogAccountStatus{
			Health:               DatadogAccountHealth(s.Health),
			ReadyForUse:          s.ReadyForUse,
			ServiceCount:         s.LogServiceCount,
			ActiveServices:       s.LogActiveServices,
			OKServices:           s.OkServices,
			DisabledServices:     s.DisabledServices,
			InactiveServices:     s.InactiveServices,
			EventCount:           s.LogEventCount,
			AnalyzedCount:        s.LogEventAnalyzedCount,
			PendingPolicyCount:   s.PolicyPendingCount,
			ApprovedPolicyCount:  s.PolicyApprovedCount,
			DismissedPolicyCount: s.PolicyDismissedCount,
		}, nil
	}

	return nil, nil
}

func (c *Client) gql(ctx context.Context) (genqlient.Client, error) {
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
			token: token,
			base:  transport,
		},
	}

	return genqlient.NewClient(c.endpoint, httpClient), nil
}

type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
