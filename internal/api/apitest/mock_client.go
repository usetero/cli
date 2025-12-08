package apitest

import (
	"context"

	"github.com/usetero/cli/pkg/client"
)

// MockClient implements api.Client for testing.
type MockClient struct {
	SetAccessTokenFunc                            func(token string)
	ListOrganizationsFunc                         func(ctx context.Context) (*client.ListOrganizationsResponse, error)
	CreateOrganizationAndBootstrapFunc            func(ctx context.Context, input client.CreateOrganizationInput) (*client.CreateOrganizationAndBootstrapResponse, error)
	ListAccountsFunc                              func(ctx context.Context, organizationID string) (*client.ListAccountsResponse, error)
	CreateAccountFunc                             func(ctx context.Context, input client.CreateAccountInput) (*client.CreateAccountResponse, error)
	GetAccountFunc                                func(ctx context.Context, accountID string) (*client.GetAccountResponse, error)
	ValidateDatadogApiKeyFunc                     func(ctx context.Context, input client.ValidateDatadogApiKeyInput) (*client.ValidateDatadogApiKeyResponse, error)
	CreateDatadogAccountWithCredentialsFunc       func(ctx context.Context, input client.CreateDatadogAccountWithCredentialsInput) (*client.CreateDatadogAccountWithCredentialsResponse, error)
	GetDatadogAccountServiceDiscoveryProgressFunc func(ctx context.Context, id string) (*client.GetDatadogAccountServiceDiscoveryProgressResponse, error)
	GetDatadogAccountLogDiscoveryProgressFunc     func(ctx context.Context, id string) (*client.GetDatadogAccountLogDiscoveryProgressResponse, error)
}

func (m *MockClient) SetAccessToken(token string) {
	if m.SetAccessTokenFunc != nil {
		m.SetAccessTokenFunc(token)
	}
}

func (m *MockClient) ListOrganizations(ctx context.Context) (*client.ListOrganizationsResponse, error) {
	if m.ListOrganizationsFunc != nil {
		return m.ListOrganizationsFunc(ctx)
	}
	return nil, nil
}

func (m *MockClient) CreateOrganizationAndBootstrap(ctx context.Context, input client.CreateOrganizationInput) (*client.CreateOrganizationAndBootstrapResponse, error) {
	if m.CreateOrganizationAndBootstrapFunc != nil {
		return m.CreateOrganizationAndBootstrapFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) ListAccounts(ctx context.Context, organizationID string) (*client.ListAccountsResponse, error) {
	if m.ListAccountsFunc != nil {
		return m.ListAccountsFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m *MockClient) CreateAccount(ctx context.Context, input client.CreateAccountInput) (*client.CreateAccountResponse, error) {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) GetAccount(ctx context.Context, accountID string) (*client.GetAccountResponse, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *MockClient) ValidateDatadogApiKey(ctx context.Context, input client.ValidateDatadogApiKeyInput) (*client.ValidateDatadogApiKeyResponse, error) {
	if m.ValidateDatadogApiKeyFunc != nil {
		return m.ValidateDatadogApiKeyFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) CreateDatadogAccountWithCredentials(ctx context.Context, input client.CreateDatadogAccountWithCredentialsInput) (*client.CreateDatadogAccountWithCredentialsResponse, error) {
	if m.CreateDatadogAccountWithCredentialsFunc != nil {
		return m.CreateDatadogAccountWithCredentialsFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) GetDatadogAccountServiceDiscoveryProgress(ctx context.Context, id string) (*client.GetDatadogAccountServiceDiscoveryProgressResponse, error) {
	if m.GetDatadogAccountServiceDiscoveryProgressFunc != nil {
		return m.GetDatadogAccountServiceDiscoveryProgressFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClient) GetDatadogAccountLogDiscoveryProgress(ctx context.Context, id string) (*client.GetDatadogAccountLogDiscoveryProgressResponse, error) {
	if m.GetDatadogAccountLogDiscoveryProgressFunc != nil {
		return m.GetDatadogAccountLogDiscoveryProgressFunc(ctx, id)
	}
	return nil, nil
}
