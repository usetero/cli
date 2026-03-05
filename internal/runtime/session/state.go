package session

import (
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// State is the current account-runtime lifecycle state.
type State struct {
	Running   bool
	AccountID tenancy.AccountID
	DBPath    sqlite.DatabasePath
}
