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
		bootstrap.State{
			User:      m.state.user,
			Org:       m.state.org,
			Account:   m.state.account,
			Workspace: m.state.workspace,
			DDSite:    m.state.ddSite,
			DDAPIKey:  m.state.ddAPIKey,
			DDAccount: m.state.ddAccount,
		},
	)
	return rewind
}
