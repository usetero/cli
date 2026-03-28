package onboarding

import "github.com/usetero/cli/internal/interfaces/tui/core"

func (m *Model) Busy() *core.Busy {
	if m.busy != nil {
		return m.busy
	}
	if m.loading {
		return &core.Busy{Label: "Loading Onboarding State"}
	}
	return m.Router.Busy()
}
