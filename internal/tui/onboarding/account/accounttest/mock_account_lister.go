package accounttest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockAccountLister is a mock implementation of account.AccountLister.
type MockAccountLister struct {
	ListFunc func(ctx context.Context, orgID string) ([]api.Account, error)
}

func (m *MockAccountLister) List(ctx context.Context, orgID string) ([]api.Account, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, orgID)
	}
	return nil, nil
}
