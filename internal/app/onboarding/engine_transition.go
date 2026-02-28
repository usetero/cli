package onboarding

import tea "charm.land/bubbletea/v2"

func (m *Model) handleTransition(msg tea.Msg) tea.Cmd {
	if out, ok := m.handlePreflightAndAuthTransition(msg); ok {
		return m.applyTransitionOutcome(out)
	}
	if out, ok := m.handleOrgAccountRuntimeTransition(msg); ok {
		return m.applyTransitionOutcome(out)
	}
	if out, ok := m.handleDatadogTransition(msg); ok {
		return m.applyTransitionOutcome(out)
	}
	if out, ok := m.handleWorkspaceSyncTransition(msg); ok {
		return m.applyTransitionOutcome(out)
	}
	return nil
}
