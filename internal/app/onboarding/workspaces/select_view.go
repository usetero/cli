package workspaces

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
)

// View renders the workspace selection UI.
func (m *SelectModel) View() string {
	s := m.theme.Styles

	if m.list.IsLoading() {
		return m.list.View()
	}

	if m.list.HasError() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			s.Error.Render(errorfmt.UserFacing(m.list.Error(), "Failed to load workspaces.")),
			s.Help.Render("Press 'r' to retry."),
		)
	}

	title := s.Title.Render("Select your workspace")
	subtitle := s.Help.Render("Workspaces organize your conversations")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.list.View(),
	)
}
