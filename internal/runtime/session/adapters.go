package session

import (
	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func toSQLiteAccountID(accountID tenancy.AccountID) sqlite.AccountID {
	return sqlite.AccountID(accountID)
}

func toSyncerAccountID(accountID tenancy.AccountID) pssyncer.AccountID {
	return pssyncer.AccountID(accountID)
}
