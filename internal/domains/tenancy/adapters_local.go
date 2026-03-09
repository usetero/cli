package tenancy

import (
	"time"

	workspacesdb "github.com/usetero/cli/internal/domains/tenancy/db/workspacesgen"
)

func ptrString(v string) *string {
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func toWorkspacesDBAccountID(id AccountID) string {
	return string(id)
}

func toWorkspacesDBWorkspaceID(id WorkspaceID) string {
	return string(id)
}

func fromWorkspacesDBWorkspace(row workspacesdb.Workspace) Workspace {
	workspace := Workspace{
		ID:        WorkspaceID(derefString(row.ID)),
		AccountID: AccountID(derefString(row.AccountID)),
		Name:      derefString(row.Name),
	}
	if t, err := time.Parse(time.RFC3339, derefString(row.CreatedAt)); err == nil {
		workspace.CreatedAt = t
	}
	return workspace
}

func fromWorkspacesDBListRow(row workspacesdb.ListByAccountRow) Workspace {
	workspace := Workspace{
		ID:        WorkspaceID(derefString(row.ID)),
		AccountID: AccountID(derefString(row.AccountID)),
		Name:      derefString(row.Name),
	}
	if t, err := time.Parse(time.RFC3339, derefString(row.CreatedAt)); err == nil {
		workspace.CreatedAt = t
	}
	return workspace
}
