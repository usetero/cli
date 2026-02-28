package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
)

// setStep sets the current step and initializes it.
func (m *Model) setStep(step StepCore) tea.Cmd {
	m.step = step
	m.step.SetSize(m.width, m.height)
	return m.step.Init()
}

func (m *Model) goToGate(gate Gate) tea.Cmd {
	requested := gate
	if rewind := m.rewindGateFor(gate); rewind != gate {
		m.scope.Warn("rewinding onboarding gate due to unmet requirements", "requested_gate", requested.String(), "rewind_gate", rewind.String())
		gate = rewind
	}
	m.gate = gate
	m.scope.Debug("onboarding gate transition", slog.String("gate", gate.String()))
	return m.runGate(gate)
}

func (m *Model) runGate(gate Gate) tea.Cmd {
	step, ok := m.newStepForGate(gate)
	if !ok {
		m.scope.Error("unsupported onboarding gate", slog.String("gate", gate.String()))
		return nil
	}
	return m.setStep(step)
}
