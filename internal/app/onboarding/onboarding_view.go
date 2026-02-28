package onboarding

import "charm.land/lipgloss/v2"

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	view := m.step.View()
	policy := m.displayPolicyForGate(m.gate)
	if policy.hidden {
		view = m.hiddenStepView(policy.status)
	}
	if v, ok := m.step.(VisibilityProvider); ok && v.Hidden() {
		view = m.hiddenStepView(v.StatusText())
	}

	// Bottom-align step content in available space
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(view)
}

func (m *Model) hiddenStepView(status string) string {
	s := m.theme.Styles
	title := s.Title.Render("Getting ready")
	body := s.Body.Render(status)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}
