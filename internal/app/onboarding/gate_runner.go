package onboarding

import tea "charm.land/bubbletea/v2"

// GateRunner constructs and initializes the step for a specific gate.
// It keeps gate-specific construction logic out of the orchestration loop.
type GateRunner interface {
	Gate() Gate
	NewStep(m *Model) Step
}

type gateRunnerFunc struct {
	gate Gate
	new  func(m *Model) Step
}

func (r gateRunnerFunc) Gate() Gate { return r.gate }

func (r gateRunnerFunc) NewStep(m *Model) Step { return r.new(m) }

func (m *Model) runGate(gate Gate) tea.Cmd {
	def := m.definitionForGate(gate)
	return m.setStep(def.runner.NewStep(m))
}
