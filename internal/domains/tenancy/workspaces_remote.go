package tenancy

import (
	"context"
	"fmt"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteWorkspaceClient interface {
	DeleteWorkspace(ctx context.Context, workspaceID controlplane.WorkspaceID) error
	ListWorkspaces(ctx context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error)
}

// RemoteWorkspaceService uses control-plane API for workspace operations.
type RemoteWorkspaceService struct {
	client remoteWorkspaceClient
}

func NewRemoteWorkspaceService(client remoteWorkspaceClient) *RemoteWorkspaceService {
	if client == nil {
		panic("tenancy remote workspace service requires client")
	}
	return &RemoteWorkspaceService{client: client}
}

func (s *RemoteWorkspaceService) Create(_ context.Context, _ WorkspaceCreate) (WorkspaceID, error) {
	return "", fmt.Errorf("tenancy remote workspace create is not implemented")
}

func (s *RemoteWorkspaceService) Delete(ctx context.Context, id WorkspaceID) error {
	if id == "" {
		return fmt.Errorf("workspace id is required")
	}
	return s.client.DeleteWorkspace(ctx, toControlPlaneWorkspaceID(id))
}

func (s *RemoteWorkspaceService) ListByAccount(ctx context.Context, accountID AccountID) ([]Workspace, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}

	rows, err := s.client.ListWorkspaces(ctx, toControlPlaneAccountID(accountID))
	if err != nil {
		return nil, err
	}

	out := make([]Workspace, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromControlPlaneWorkspace(row, accountID))
	}
	return out, nil
}
