package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
)

// MockClient implements api.Client for testing.
type MockClient struct {
	SetAccountIDFunc                        func(accountID domain.AccountID)
	RawQueryFunc                            func(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error)
	ListOrganizationsFunc                   func(ctx context.Context) (*gen.ListOrganizationsResponse, error)
	CreateOrganizationAndBootstrapFunc      func(ctx context.Context, input gen.CreateOrganizationInput) (*gen.CreateOrganizationAndBootstrapResponse, error)
	ListAccountsFunc                        func(ctx context.Context, organizationID string) (*gen.ListAccountsResponse, error)
	CreateAccountFunc                       func(ctx context.Context, input gen.CreateAccountInput) (*gen.CreateAccountResponse, error)
	GetAccountFunc                          func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error)
	ValidateDatadogApiKeyFunc               func(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error)
	CreateDatadogAccountWithCredentialsFunc func(ctx context.Context, input gen.CreateDatadogAccountWithCredentialsInput) (*gen.CreateDatadogAccountWithCredentialsResponse, error)
	GetDatadogAccountStatusFunc             func(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error)
	ListWorkspacesFunc                      func(ctx context.Context, accountID string) (*gen.ListWorkspacesResponse, error)
	CreateConversationFunc                  func(ctx context.Context, input gen.CreateConversationInput) (*gen.CreateConversationResponse, error)
	UpdateConversationFunc                  func(ctx context.Context, id string, input gen.UpdateConversationInput) (*gen.UpdateConversationResponse, error)
	DeleteConversationFunc                  func(ctx context.Context, id string) (*gen.DeleteConversationResponse, error)
	CreateMessageFunc                       func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error)
	EnableServiceFunc                       func(ctx context.Context, serviceID string) (*gen.EnableServiceResponse, error)
	DisableServiceFunc                      func(ctx context.Context, serviceID string) (*gen.DisableServiceResponse, error)
	ApproveLogEventPolicyFunc               func(ctx context.Context, id string) (*gen.ApproveLogEventPolicyResponse, error)
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

func (m *MockClient) CreateOrganizationAndBootstrap(ctx context.Context, input gen.CreateOrganizationInput) (*gen.CreateOrganizationAndBootstrapResponse, error) {
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

func (m *MockClient) CreateAccount(ctx context.Context, input gen.CreateAccountInput) (*gen.CreateAccountResponse, error) {
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

func (m *MockClient) CreateDatadogAccountWithCredentials(ctx context.Context, input gen.CreateDatadogAccountWithCredentialsInput) (*gen.CreateDatadogAccountWithCredentialsResponse, error) {
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

func (m *MockClient) ListWorkspaces(ctx context.Context, accountID string) (*gen.ListWorkspacesResponse, error) {
	if m.ListWorkspacesFunc != nil {
		return m.ListWorkspacesFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *MockClient) CreateConversation(ctx context.Context, input gen.CreateConversationInput) (*gen.CreateConversationResponse, error) {
	if m.CreateConversationFunc != nil {
		return m.CreateConversationFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockClient) UpdateConversation(ctx context.Context, id string, input gen.UpdateConversationInput) (*gen.UpdateConversationResponse, error) {
	if m.UpdateConversationFunc != nil {
		return m.UpdateConversationFunc(ctx, id, input)
	}
	return nil, nil
}

func (m *MockClient) DeleteConversation(ctx context.Context, id string) (*gen.DeleteConversationResponse, error) {
	if m.DeleteConversationFunc != nil {
		return m.DeleteConversationFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClient) CreateMessage(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
	if m.CreateMessageFunc != nil {
		return m.CreateMessageFunc(ctx, input)
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

func (m *MockClient) ApproveLogEventPolicy(ctx context.Context, id string) (*gen.ApproveLogEventPolicyResponse, error) {
	if m.ApproveLogEventPolicyFunc != nil {
		return m.ApproveLogEventPolicyFunc(ctx, id)
	}
	return nil, nil
}
