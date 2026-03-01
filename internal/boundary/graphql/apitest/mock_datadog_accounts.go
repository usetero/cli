package apitest

import (
	"context"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
)

// MockDatadogAccounts implements graphql.DatadogAccounts for testing.
type MockDatadogAccounts struct {
	HasAccountFunc     func(ctx context.Context, accountID domain.AccountID) (bool, error)
	GetAccountFunc     func(ctx context.Context, accountID domain.AccountID) (*graphql.DatadogAccount, error)
	ValidateAPIKeyFunc func(ctx context.Context, input graphql.ValidateAPIKeyInput) (bool, string, error)
	CreateAccountFunc  func(ctx context.Context, input graphql.CreateDatadogAccountInput) (*graphql.DatadogAccount, error)
	GetStatusFunc      func(ctx context.Context, datadogAccountID domain.DatadogAccountID) (*graphql.DatadogAccountStatus, error)
}

// NewMockDatadogAccounts creates a MockDatadogAccounts with sensible defaults.
func NewMockDatadogAccounts() *MockDatadogAccounts {
	return &MockDatadogAccounts{}
}

func (m *MockDatadogAccounts) HasAccount(ctx context.Context, accountID domain.AccountID) (bool, error) {
	if m.HasAccountFunc != nil {
		return m.HasAccountFunc(ctx, accountID)
	}
	return false, nil
}

func (m *MockDatadogAccounts) GetAccount(ctx context.Context, accountID domain.AccountID) (*graphql.DatadogAccount, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *MockDatadogAccounts) ValidateAPIKey(ctx context.Context, input graphql.ValidateAPIKeyInput) (bool, string, error) {
	if m.ValidateAPIKeyFunc != nil {
		return m.ValidateAPIKeyFunc(ctx, input)
	}
	return false, "", nil
}

func (m *MockDatadogAccounts) CreateAccount(ctx context.Context, input graphql.CreateDatadogAccountInput) (*graphql.DatadogAccount, error) {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockDatadogAccounts) GetStatus(ctx context.Context, datadogAccountID domain.DatadogAccountID) (*graphql.DatadogAccountStatus, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, datadogAccountID)
	}
	return nil, nil
}
