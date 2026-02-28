package datadog

import "charm.land/lipgloss/v2"

// View renders the App key entry UI.
func (m *AppKeyModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Enter your Datadog Application Key")
	subtitle := s.Help.Render("You can find this in Datadog under Organization Settings → Application Keys")

	var status string
	if m.creating {
		status = s.Help.Render("Connecting to Datadog...")
	} else if m.err != nil {
		status = s.Error.Render(appKeyErrorMessage(m.err))
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
