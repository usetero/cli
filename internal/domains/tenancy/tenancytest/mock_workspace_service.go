package tenancytest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// MockWorkspaceService is a functional mock for tenancy.WorkspaceService.
type MockWorkspaceService struct {
	CreateFn        func(ctx context.Context, create tenancy.WorkspaceCreate) (tenancy.WorkspaceID, error)
	DeleteFn        func(ctx context.Context, id tenancy.WorkspaceID) error
	ListByAccountFn func(ctx context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error)
}

var _ tenancy.WorkspaceService = (*MockWorkspaceService)(nil)

func (m *MockWorkspaceService) Create(ctx context.Context, create tenancy.WorkspaceCreate) (tenancy.WorkspaceID, error) {
	if m.CreateFn == nil {
		return "", nil
	}
	return m.CreateFn(ctx, create)
}

func (m *MockWorkspaceService) Delete(ctx context.Context, id tenancy.WorkspaceID) error {
	if m.DeleteFn == nil {
		return nil
	}
	return m.DeleteFn(ctx, id)
}

func (m *MockWorkspaceService) ListByAccount(ctx context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
	if m.ListByAccountFn == nil {
		return nil, nil
	}
	return m.ListByAccountFn(ctx, accountID)
}
