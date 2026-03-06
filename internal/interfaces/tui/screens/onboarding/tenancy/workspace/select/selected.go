package workspaceselect

import "github.com/usetero/cli/internal/domains/tenancy"

// SelectedMsg reports that the user confirmed a workspace selection.
type SelectedMsg struct {
	WorkspaceID tenancy.WorkspaceID
}
