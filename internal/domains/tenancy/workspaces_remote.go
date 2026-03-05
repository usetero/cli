package tenancy

import (
	"context"
	"fmt"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteWorkspaceClient interface {
	ListWorkspaces(ctx context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error)
}

// RemoteWorkspaceService uses control-plane API for workspace operations.
type RemoteWorkspaceService struct {
	client remoteWorkspaceClient
}

func NewRemoteWorkspaceService(client remoteWorkspaceClient) *RemoteWorkspaceService {
	return &RemoteWorkspaceService{client: client}
}

func (s *RemoteWorkspaceService) Create(_ context.Context, _ AccountID, _ string) (WorkspaceID, error) {
	return "", fmt.Errorf("tenancy remote workspace create is not implemented")
}

func (s *RemoteWorkspaceService) Delete(_ context.Context, _ WorkspaceID) error {
	return fmt.Errorf("tenancy remote workspace delete is not implemented")
}

func (s *RemoteWorkspaceService) ListByAccount(ctx context.Context, accountID AccountID) ([]Workspace, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("tenancy remote workspace service is not initialized")
	}
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
