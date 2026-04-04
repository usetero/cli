package commandbar

func (m *Model) CapturingInput() bool {
	if m.paletteOpen {
		return true
	}
	if m.err != nil || m.busy != nil {
		return false
	}
	if m.mode == ModeAction {
		return true
	}
	return m.children.action.CapturingInput()
}
