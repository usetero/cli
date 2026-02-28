package datadog

import "github.com/usetero/cli/internal/app/onboarding/errorfmt"

// View renders the check UI.
func (m *CheckModel) View() string {
	s := m.theme.Styles

	if m.err != nil {
		return s.Error.Render(errorfmt.UserFacing(m.err, "Failed to check Datadog configuration. Press 'r' to retry."))
	}
	return s.Title.Render("Checking Datadog configuration...")
}
