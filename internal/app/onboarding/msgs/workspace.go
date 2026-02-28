package msgs

import "github.com/usetero/cli/internal/domain"

// WorkspaceSelected is emitted when user selects a workspace.
type WorkspaceSelected struct {
	Workspace domain.Workspace
}
