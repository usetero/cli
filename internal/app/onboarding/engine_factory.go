package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
)

// setStep sets the current step and initializes it.
func (m *Model) setStep(step Step) tea.Cmd {
	m.step = step
	m.step.SetSize(m.width, m.height)
	return m.step.Init()
}

func (m *Model) goToGate(gate Gate) tea.Cmd {
	m.gate = gate
	m.scope.Debug("onboarding gate transition", slog.String("gate", gate.String()))
	if rewind := m.rewindGateFor(gate); rewind != gate {
		m.scope.Warn("rewinding onboarding gate due to unmet requirements", "requested_gate", gate.String(), "rewind_gate", rewind.String())
		gate = rewind
	}
	return m.runGate(gate)
}
