package datadog

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
)

// View renders the region selection UI.
func (m *RegionModel) View() string {
	s := m.theme.Styles
	colors := m.theme

	title := s.Title.Render("Select your Datadog region")
	subtitle := s.Help.Render("Choose the region where your Datadog account is hosted")

	var optionViews []string
	for i, r := range domain.DatadogRegions {
		var view string
		if i == m.selected {
			nameStyle := lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg).Bold(true)
			view = nameStyle.Render("> " + r.Name)
		} else {
			view = s.Body.Render("  " + r.Name)
		}
		optionViews = append(optionViews, view)
	}

	parts := []string{title, subtitle, ""}
	parts = append(parts, optionViews...)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
