package workspaceselect

import "github.com/usetero/cli/internal/domains/tenancy"

// SelectedMsg reports a workspace choice from the select page.
type SelectedMsg struct {
	WorkspaceID tenancy.WorkspaceID
}
