package datadog

import "charm.land/lipgloss/v2"

// View renders the API key entry UI.
func (m *APIKeyModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Enter your Datadog API Key")
	subtitle := s.Help.Render("You can find this in Datadog under Organization Settings → API Keys")

	var status string
	if m.validating {
		status = s.Help.Render("Validating...")
	} else if m.err != nil {
		status = s.Error.Render(m.err.Error())
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
