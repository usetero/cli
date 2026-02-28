package onboarding

import "charm.land/lipgloss/v2"

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	view := m.step.View()
	if v, ok := m.step.(VisibilityProvider); ok && v.Hidden() {
		status := defaultStatusForGate(m.gate)
		if s, ok := m.step.(StatusProvider); ok {
			status = s.Status()
		} else {
			status.Details = v.StatusText()
		}
		view = m.hiddenStepView(status)
	}

	// Bottom-align step content in available space
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(view)
}

func (m *Model) hiddenStepView(status appmsg.StepStatus) string {
	s := m.theme.Styles
	title := s.Title.Render(status.Title)
	body := s.Body.Render(status.Details)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}
