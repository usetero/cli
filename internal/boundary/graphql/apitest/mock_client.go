package apitest

import (
	"context"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
)

// MockClient implements graphql.Client for testing.
type MockClient struct {
	SetAccountIDFunc                        func(accountID domain.AccountID)
	WithAccountIDFunc                       func(accountID domain.AccountID) graphql.Client
	RawQueryFunc                            func(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error)
	ListOrganizationsFunc                   func(ctx context.Context) (*gen.ListOrganizationsResponse, error)
	CreateOrganizationAndBootstrapFunc      func(ctx context.Context, input gen.OrganizationCreateInput) (*gen.CreateOrganizationAndBootstrapResponse, error)
	ListAccountsFunc                        func(ctx context.Context, organizationID string) (*gen.ListAccountsResponse, error)
	CreateAccountFunc                       func(ctx context.Context, input gen.AccountCreateInput) (*gen.CreateAccountResponse, error)
	GetAccountFunc                          func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error)
	ValidateDatadogApiKeyFunc               func(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error)
	CreateDatadogAccountWithCredentialsFunc func(ctx context.Context, input gen.DatadogAccountCreateInput) (*gen.CreateDatadogAccountWithCredentialsResponse, error)
	GetDatadogAccountStatusFunc             func(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error)
	EnableServiceFunc                       func(ctx context.Context, serviceID string) (*gen.EnableServiceResponse, error)
	DisableServiceFunc                      func(ctx context.Context, serviceID string) (*gen.DisableServiceResponse, error)
	GetIssueSummaryFunc                     func(ctx context.Context) (*gen.GetIssueSummaryResponse, error)
	ListChecksFunc                          func(ctx context.Context) (*gen.ListChecksResponse, error)
	ListEdgeInstancesFunc                   func(ctx context.Context) (*gen.ListEdgeInstancesResponse, error)
}

// NewMockClient creates a MockClient with sensible defaults.
func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) SetAccountID(accountID domain.AccountID) {
	if m.SetAccountIDFunc != nil {
		m.SetAccountIDFunc(accountID)
	}
}

func (m *MockClient) WithAccountID(accountID domain.AccountID) graphql.Client {
	if m.WithAccountIDFunc != nil {
		return m.WithAccountIDFunc(accountID)
	}
	if m.SetAccountIDFunc != nil {
		m.SetAccountIDFunc(accountID)
	}
	return m
}

func (m *MockClient) RawQuery(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	if m.RawQueryFunc != nil {
		return m.RawQueryFunc(ctx, query, variables)
	}
	return nil, nil
}

func (m *MockClient) ListOrganizations(ctx context.Context) (*gen.ListOrganizationsResponse, error) {
	if m.ListOrganizationsFunc != nil {
		return m.ListOrganizationsFunc(ctx)
	}
	return nil, nil
}

func (m *MockClient) CreateOrganizationAndBootstrap(ctx context.Context, input gen.OrganizationCreateInput) (*gen.CreateOrganizationAndBootstrapResponse, error) {
	if m.CreateOrganizationAndBootstrapFunc != nil {
		return m.CreateOrganizationAndBootstrapFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) ListAccounts(ctx context.Context, organizationID string) (*gen.ListAccountsResponse, error) {
	if m.ListAccountsFunc != nil {
		return m.ListAccountsFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m *MockClient) CreateAccount(ctx context.Context, input gen.AccountCreateInput) (*gen.CreateAccountResponse, error) {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) GetAccount(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *MockClient) ValidateDatadogApiKey(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error) {
	if m.ValidateDatadogApiKeyFunc != nil {
		return m.ValidateDatadogApiKeyFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) CreateDatadogAccountWithCredentials(ctx context.Context, input gen.DatadogAccountCreateInput) (*gen.CreateDatadogAccountWithCredentialsResponse, error) {
	if m.CreateDatadogAccountWithCredentialsFunc != nil {
		return m.CreateDatadogAccountWithCredentialsFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) GetDatadogAccountStatus(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error) {
	if m.GetDatadogAccountStatusFunc != nil {
		return m.GetDatadogAccountStatusFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClient) EnableService(ctx context.Context, serviceID string) (*gen.EnableServiceResponse, error) {
	if m.EnableServiceFunc != nil {
		return m.EnableServiceFunc(ctx, serviceID)
	}
	return nil, nil
}

func (m *MockClient) DisableService(ctx context.Context, serviceID string) (*gen.DisableServiceResponse, error) {
	if m.DisableServiceFunc != nil {
		return m.DisableServiceFunc(ctx, serviceID)
	}
	return nil, nil
}

func (m *MockClient) GetIssueSummary(ctx context.Context) (*gen.GetIssueSummaryResponse, error) {
	if m.GetIssueSummaryFunc != nil {
		return m.GetIssueSummaryFunc(ctx)
	}
	return nil, nil
}

func (m *MockClient) ListChecks(ctx context.Context) (*gen.ListChecksResponse, error) {
	if m.ListChecksFunc != nil {
		return m.ListChecksFunc(ctx)
	}
	return nil, nil
}

func (m *MockClient) ListEdgeInstances(ctx context.Context) (*gen.ListEdgeInstancesResponse, error) {
	if m.ListEdgeInstancesFunc != nil {
		return m.ListEdgeInstancesFunc(ctx)
	}
	return nil, nil
}
