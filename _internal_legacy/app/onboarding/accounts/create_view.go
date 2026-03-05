package accounts

import (
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

// View renders the account creation UI.
func (m *CreateModel) View() string {
	return stepkit.RenderCreateForm(
		m.theme,
		"Create a Datadog account",
		"Accounts connect to a Datadog organization",
		m.input.View(),
		m.creating,
		m.err,
		"Failed to create account. Try again.",
	)
}
