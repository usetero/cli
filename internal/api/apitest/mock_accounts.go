package apitest

import (
	"context"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
)

// MockAccounts implements api.Accounts for testing.
type MockAccounts struct {
	ListFunc   func(ctx context.Context, organizationID string) ([]api.Account, error)
	GetFunc    func(ctx context.Context, accountID string) (*api.Account, error)
	CreateFunc func(ctx context.Context, id uuid.UUID, organizationID, name string) (*api.Account, error)
}

func (m *MockAccounts) List(ctx context.Context, organizationID string) ([]api.Account, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m *MockAccounts) Get(ctx context.Context, accountID string) (*api.Account, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *MockAccounts) Create(ctx context.Context, id uuid.UUID, organizationID, name string) (*api.Account, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, id, organizationID, name)
	}
	return nil, nil
}
