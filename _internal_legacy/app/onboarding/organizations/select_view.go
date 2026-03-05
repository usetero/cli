package organizations

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
)

// View renders the organization selection UI.
func (m *SelectModel) View() string {
	s := m.theme.Styles

	if m.list.IsLoading() {
		return m.list.View()
	}

	if m.refreshingToken && m.refreshLoader != nil {
		return m.refreshLoader.View()
	}

	if m.list.HasError() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			s.Error.Render(errorfmt.UserFacing(m.list.Error(), "Failed to load organizations.")),
			s.Help.Render("Press 'r' to retry."),
		)
	}

	title := s.Title.Render("Select your organization")
	subtitle := s.Help.Render("Press 'n' to create a new organization")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.list.View(),
	)
}
