package apitest

import (
	"context"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
)

// MockAccounts implements graphql.Accounts for testing.
type MockAccounts struct {
	ListFunc   func(ctx context.Context, organizationID domain.OrganizationID) ([]domain.Account, error)
	GetFunc    func(ctx context.Context, accountID domain.AccountID) (*domain.Account, error)
	CreateFunc func(ctx context.Context, input graphql.CreateAccountInput) (*domain.Account, error)
}

// NewMockAccounts creates a MockAccounts with sensible defaults.
func NewMockAccounts() *MockAccounts {
	return &MockAccounts{}
}

func (m *MockAccounts) List(ctx context.Context, organizationID domain.OrganizationID) ([]domain.Account, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m *MockAccounts) Get(ctx context.Context, accountID domain.AccountID) (*domain.Account, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *MockAccounts) Create(ctx context.Context, input graphql.CreateAccountInput) (*domain.Account, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}
