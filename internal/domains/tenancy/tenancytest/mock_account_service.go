package tenancytest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// MockAccountService is a functional mock for tenancy.AccountService.
type MockAccountService struct {
	CreateFn func(ctx context.Context, create tenancy.AccountCreate) (tenancy.AccountID, error)
	DeleteFn func(ctx context.Context, id tenancy.AccountID) error
	ListFn   func(ctx context.Context) ([]tenancy.Account, error)
}

var _ tenancy.AccountService = (*MockAccountService)(nil)

func NewMockAccountService() *MockAccountService {
	return &MockAccountService{}
}

func (m *MockAccountService) Create(ctx context.Context, create tenancy.AccountCreate) (tenancy.AccountID, error) {
	if m.CreateFn == nil {
		return "", nil
	}
	return m.CreateFn(ctx, create)
}

func (m *MockAccountService) Delete(ctx context.Context, id tenancy.AccountID) error {
	if m.DeleteFn == nil {
		return nil
	}
	return m.DeleteFn(ctx, id)
}

func (m *MockAccountService) List(ctx context.Context) ([]tenancy.Account, error) {
	if m.ListFn == nil {
		return nil, nil
	}
	return m.ListFn(ctx)
}
