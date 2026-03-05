package onboarding

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
)

func (m *Model) applyTransition(event bootstrap.Event, msg tea.Msg) bootstrap.Transition {
	if preflight, ok := msg.(bootstrap.PreflightResolved); ok {
		m.logPreflightResolved(preflight)
	}

	transition := bootstrap.ApplyEvent(m.state, event)
	m.state = transition.State
	m.syncServicesToState()
	return transition
}
