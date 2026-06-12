package graphql

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Workspaces provides access to workspaces.
type Workspaces interface {
	List(ctx context.Context, accountID domain.AccountID) ([]domain.Workspace, error)
}

// WorkspaceService handles workspace-related API operations.
type WorkspaceService struct {
	client Client
	scope  log.Scope
}

// Ensure WorkspaceService implements Workspaces.
var _ Workspaces = (*WorkspaceService)(nil)

// NewWorkspaceService creates a new workspace service.
func NewWorkspaceService(client Client, scope log.Scope) *WorkspaceService {
	return &WorkspaceService{
		client: client,
		scope:  scope.Child("workspaces"),
	}
}

// List returns the workspaces for an account.
//
// TODO(drop-powersync): workspaces were removed from the control plane — the
// account is the working context now. As an interim, this returns a single
// synthetic workspace mirroring the account so the onboarding selection step is
// a no-op auto-select. The full workspace→account rename is task #7.
func (s *WorkspaceService) List(_ context.Context, accountID domain.AccountID) ([]domain.Workspace, error) {
	return []domain.Workspace{
		{ID: domain.WorkspaceID(accountID.String()), Name: "Default"},
	}, nil
}
