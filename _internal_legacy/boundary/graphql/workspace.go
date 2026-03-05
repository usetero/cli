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

// List fetches all workspaces for an account.
func (s *WorkspaceService) List(ctx context.Context, accountID domain.AccountID) ([]domain.Workspace, error) {
	s.scope.Debug("fetching workspaces from API", "accountID", accountID)
	resp, err := s.client.ListWorkspaces(ctx, accountID.String())
	if err != nil {
		s.scope.Error("failed to fetch workspaces", "error", err, "accountID", accountID)
		return nil, err
	}

	// Convert GraphQL response to domain model
	workspaces := make([]domain.Workspace, len(resp.Workspaces.Edges))
	for i, edge := range resp.Workspaces.Edges {
		workspaces[i] = domain.Workspace{
			ID:   domain.WorkspaceID(edge.Node.Id),
			Name: edge.Node.Name,
		}
	}

	s.scope.Debug("fetched workspaces from API", "count", len(workspaces))
	return workspaces, nil
}
