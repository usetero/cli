package tenancy

import (
	"context"
	"fmt"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteWorkspaceClient interface {
	DeleteWorkspace(ctx context.Context, workspaceID controlplane.WorkspaceID) error
	ListWorkspaces(ctx context.Context) ([]controlplane.Workspace, error)
}

// RemoteWorkspaceService uses control-plane API for workspace operations.
type RemoteWorkspaceService struct {
	client    remoteWorkspaceClient
	accountID AccountID
}

func NewRemoteWorkspaceService(client remoteWorkspaceClient, accountID AccountID) *RemoteWorkspaceService {
	if client == nil {
		panic("tenancy remote workspace service requires client")
	}
	if accountID == "" {
		panic("tenancy remote workspace service requires account id")
	}
	return &RemoteWorkspaceService{client: client, accountID: accountID}
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

func (s *RemoteWorkspaceService) List(ctx context.Context) ([]Workspace, error) {
	rows, err := s.client.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Workspace, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromControlPlaneWorkspace(row, s.accountID))
	}
	return out, nil
}
