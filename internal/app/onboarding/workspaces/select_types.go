package workspaces

import "github.com/usetero/cli/internal/domain"

func findWorkspaceByID(workspaces []domain.Workspace, id domain.WorkspaceID) *domain.Workspace {
	if id == "" {
		return nil
	}
	for _, ws := range workspaces {
		if ws.ID == id {
			resolved := ws
			return &resolved
		}
	}
	return nil
}
