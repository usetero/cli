package app

import "github.com/usetero/cli/internal/sqlite"

// openDatabase opens the SQLite database for the given account.
func (m *Model) openDatabase(accountID string) error {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			m.scope.Warn("failed to close previous database", "error", err)
		}
		m.db = nil
	}

	dbPath, err := m.storage.DatabasePath(accountID)
	if err != nil {
		return err
	}

	db, err := sqlite.Open(m.ctx, dbPath)
	if err != nil {
		return err
	}

	m.db = db
	m.scope.Info("database opened", "path", dbPath)
	return nil
}
