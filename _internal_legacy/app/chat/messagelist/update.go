package messagelist

import tea "charm.land/bubbletea/v2"

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.handleKeyPress(msg)

	case tea.MouseClickMsg:
		m.handleMouseClick(msg)

	case tea.MouseMotionMsg:
		m.handleMouseMotion(msg)

	case tea.MouseReleaseMsg:
		cmds = append(cmds, m.handleMouseRelease(msg)...)

	case tea.MouseWheelMsg:
		m.handleMouseWheel(msg)

	default:
		if lifecycleCmds, handled := m.handleLifecycle(msg); handled {
			cmds = append(cmds, lifecycleCmds...)
			return tea.Batch(cmds...)
		}
	}

	// Forward to all rounds
	for _, r := range m.rounds {
		if cmd := r.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}
