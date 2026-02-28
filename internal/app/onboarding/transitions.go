package onboarding

import tea "charm.land/bubbletea/v2"

func (m *Model) handleTransition(msg tea.Msg) tea.Cmd {
	if cmd, ok := m.transitionCmdFor(msg); ok {
		return cmd
	}
	return nil
}
