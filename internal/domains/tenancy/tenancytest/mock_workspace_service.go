package tenancytest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// MockWorkspaceService is a functional mock for tenancy.WorkspaceService.
type MockWorkspaceService struct {
	CreateFn func(ctx context.Context, create tenancy.WorkspaceCreate) (tenancy.WorkspaceID, error)
	DeleteFn func(ctx context.Context, id tenancy.WorkspaceID) error
	ListFn   func(ctx context.Context) ([]tenancy.Workspace, error)
}

var _ tenancy.WorkspaceService = (*MockWorkspaceService)(nil)

func NewMockWorkspaceService() *MockWorkspaceService {
	return &MockWorkspaceService{}
}

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

func (m *MockWorkspaceService) List(ctx context.Context) ([]tenancy.Workspace, error) {
	if m.ListFn == nil {
		return nil, nil
	}
	return m.ListFn(ctx)
}
