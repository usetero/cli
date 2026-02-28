package onboarding

import "github.com/usetero/cli/internal/core/bootstrap"

type gateRequirement = bootstrap.GateRequirement

func (m *Model) rewindGateFor(target Gate) Gate {
	def, ok := m.definitions[target]
	if !ok {
		return target
	}
	rewind := bootstrap.RewindGate(
		target,
		def.requirement,
		m.state,
	)
	return rewind
}
