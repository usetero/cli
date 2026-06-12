package app

func (m *Model) shutdown() {
	db := m.db

	if m.sessionCancel != nil {
		m.sessionCancel()
		m.sessionCancel = nil
	}
	m.sessionCtx = nil
	if m.syncer != nil {
		m.syncer.Stop()
	}
	if db != nil {
		if err := db.Close(); err != nil {
			m.scope.Warn("failed to close database", "error", err)
		}
	}
	m.db = nil
}
