package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockWorkspaces implements api.Workspaces for testing.
type MockWorkspaces struct {
	ListFunc func(ctx context.Context, accountID string) ([]api.Workspace, error)
}

// NewMockWorkspaces creates a MockWorkspaces with sensible defaults.
func NewMockWorkspaces() *MockWorkspaces {
	return &MockWorkspaces{
		ListFunc: func(ctx context.Context, accountID string) ([]api.Workspace, error) {
			return []api.Workspace{}, nil
		},
	}
}

func (m *MockWorkspaces) List(ctx context.Context, accountID string) ([]api.Workspace, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, accountID)
	}
	return nil, nil
}
