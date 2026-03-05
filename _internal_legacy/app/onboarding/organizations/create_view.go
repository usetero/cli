package organizations

import (
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

// View renders the organization creation UI.
func (m *CreateModel) View() string {
	return stepkit.RenderCreateForm(
		m.theme,
		"Create an organization",
		"Organizations contain accounts and team members",
		m.input.View(),
		m.creating,
		m.err,
		"Failed to create organization. Try again.",
	)
}
