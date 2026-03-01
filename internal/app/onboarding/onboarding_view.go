package onboarding

import "charm.land/lipgloss/v2"

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	view := m.step.View()
	if m.step.Hidden() {
		view = m.hiddenStepView(m.step.Status())
	}

	// Bottom-align step content in available space
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(view)
}

func (m *Model) hiddenStepView(status onbstatus.StepStatus) string {
	s := m.theme.Styles
	title := s.Title.Render(status.Title)
	body := s.Body.Render(status.Details)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}
