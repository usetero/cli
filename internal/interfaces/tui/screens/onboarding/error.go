package onboarding

import "github.com/usetero/cli/internal/interfaces/tui/core"

func (m *Model) Error() *core.Error {
	if m.loadErr != nil {
		return &core.Error{
			Message: m.loadErr.Error(),
			Action:  "Retry",
		}
	}
	return m.Router.Error()
}
