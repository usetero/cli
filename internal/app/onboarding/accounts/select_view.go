package accounts

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
)

// View renders the account selection UI.
func (m *SelectModel) View() string {
	s := m.theme.Styles

	if m.list.IsLoading() {
		return m.list.View()
	}

	if m.list.HasError() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			s.Error.Render(errorfmt.UserFacing(m.list.Error(), "Failed to load accounts.")),
			s.Help.Render("Press 'r' to retry."),
		)
	}

	title := s.Title.Render("Select your Datadog account")
	subtitle := s.Help.Render("Press 'n' to connect a new Datadog account")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.list.View(),
	)
}
