package onboarding

import tea "charm.land/bubbletea/v2"

func (m *Model) handleTransition(msg tea.Msg) tea.Cmd {
	if out, ok := m.transitionOutcomeFor(msg); ok {
		return m.applyTransitionOutcome(out)
	}
	return nil
}
