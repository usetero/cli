package onboarding

import "github.com/usetero/cli/internal/core/bootstrap"

func (m *Model) rewindGateFor(target Gate) Gate {
	rewind := bootstrap.RewindGate(
		target,
		bootstrap.RequirementForGate(target),
		m.state,
	)
	return rewind
}
