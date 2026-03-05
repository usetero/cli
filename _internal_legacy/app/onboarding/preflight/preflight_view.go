package preflight

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	s := m.theme.Styles
	title := s.Title.Render("Getting ready")
	statusLine := m.spinner.View() + " " + s.Body.Render(m.stageText())
	return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
}

func (m *Model) stageText() string {
	switch m.stage {
	case stageAuth:
		return "Checking authentication"
	case stageOrganizations:
		return "Loading organizations"
	case stageAccounts:
		return "Loading accounts"
	case stageFinalizing:
		return "Finalizing setup"
	default:
		return "Preparing onboarding..."
	}
}
