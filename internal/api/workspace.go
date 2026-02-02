package api

import (
	"context"

	"github.com/usetero/cli/internal/log"
)

// Workspace is the domain model for a workspace.
type Workspace struct {
	ID   string
	Name string
}

// Workspaces provides access to workspaces.
type Workspaces interface {
	List(ctx context.Context, accountID string) ([]Workspace, error)
}

// WorkspaceService handles workspace-related API operations.
type WorkspaceService struct {
	client Client
	logger log.Logger
}

// Ensure WorkspaceService implements Workspaces.
var _ Workspaces = (*WorkspaceService)(nil)

// NewWorkspaceService creates a new workspace service.
func NewWorkspaceService(client Client, logger log.Logger) *WorkspaceService {
	return &WorkspaceService{
		client: client,
		logger: logger,
	}
}

// List fetches all workspaces for an account.
func (s *WorkspaceService) List(ctx context.Context, accountID string) ([]Workspace, error) {
	s.logger.Debug("fetching workspaces from API", "accountID", accountID)
	resp, err := s.client.ListWorkspaces(ctx, accountID)
	if err != nil {
		s.logger.Error("failed to fetch workspaces", "error", err, "accountID", accountID)
		return nil, err
	}

	// Convert GraphQL response to domain model
	workspaces := make([]Workspace, len(resp.Workspaces.Edges))
	for i, edge := range resp.Workspaces.Edges {
		workspaces[i] = Workspace{
			ID:   edge.Node.Id,
			Name: edge.Node.Name,
		}
	}

	s.logger.Debug("fetched workspaces from API", "count", len(workspaces))
	return workspaces, nil
}
