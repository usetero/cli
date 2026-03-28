package onboarding

import (
	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

func (m *Model) ShortHelp() []key.Binding {
	if m.loading || m.busy != nil {
		return nil
	}
	return m.Router.ShortHelp()
}

func (m *Model) Input() *core.Input {
	if m.loading || m.busy != nil {
		return &core.Input{Label: "Loading onboarding..."}
	}
	if input := m.Router.Input(); input != nil {
		return input
	}
	return m.placeholder
}
