package accounttest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockAccountCreator is a mock implementation of account.AccountCreator.
type MockAccountCreator struct {
	CreateFunc func(ctx context.Context, orgID string, name string) (*api.Account, error)
}

func (m *MockAccountCreator) Create(ctx context.Context, orgID string, name string) (*api.Account, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, orgID, name)
	}
	return nil, nil
}
