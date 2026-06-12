package app

func (m *Model) shutdown() {
	if m.sessionCancel != nil {
		m.sessionCancel()
		m.sessionCancel = nil
	}
	m.sessionCtx = nil
}
