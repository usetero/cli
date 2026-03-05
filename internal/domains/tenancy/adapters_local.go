package tenancy

import (
	"time"

	accountsdb "github.com/usetero/cli/internal/domains/tenancy/db/accountsgen"
	workspacesdb "github.com/usetero/cli/internal/domains/tenancy/db/workspacesgen"
)

func toAccountsDBAccountID(id AccountID) string {
	return string(id)
}

func toWorkspacesDBAccountID(id AccountID) string {
	return string(id)
}

func toWorkspacesDBWorkspaceID(id WorkspaceID) string {
	return string(id)
}

func fromAccountsDBAccount(row accountsdb.Account) Account {
	account := Account{
		ID:   AccountID(row.ID),
		Name: row.Name,
	}
	if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
		account.CreatedAt = t
	}
	return account
}

func fromWorkspacesDBWorkspace(row workspacesdb.Workspace) Workspace {
	workspace := Workspace{
		ID:        WorkspaceID(row.ID),
		AccountID: AccountID(row.AccountID),
		Name:      row.Name,
	}
	if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
		workspace.CreatedAt = t
	}
	return workspace
}
