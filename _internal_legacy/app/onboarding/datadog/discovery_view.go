package datadog

import (
	"fmt"

	"charm.land/lipgloss/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

// View renders the discovery UI.
func (m *DiscoveryModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Discovering Datadog services")

	if m.err != nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			s.Error.Render(fmt.Sprintf("Discovery failed: %v", m.err)),
			"",
			s.Help.Render("Press 'r' to retry"),
		)
	}

	if m.status == nil {
		statusLine := m.spinner.View() + " " + s.Body.Render("Connecting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
	}

	if m.status.ReadyForUse {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			s.Success.Render("Discovery complete!"),
		)
	}

	statusText := m.statusText()
	statusLine := m.spinner.View() + " " + s.Body.Render(statusText)

	// Derive progress from analyzed/total events.
	var pct float64
	if m.status.EventCount > 0 {
		pct = float64(m.status.AnalyzedCount) / float64(m.status.EventCount) * 100
	}
	progressBar := m.progress.ViewAs(pct)

	parts := []string{title, "", statusLine, "", progressBar}

	if m.status.ServiceCount > 0 {
		countText := fmt.Sprintf("%d / %d services OK", m.status.OkServices, m.status.ActiveServices)
		if m.status.EventCount > 0 {
			countText += fmt.Sprintf(" · %d / %d events analyzed", m.status.AnalyzedCount, m.status.EventCount)
		}
		parts = append(parts, "", s.Help.Render(countText))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *DiscoveryModel) statusText() string {
	switch m.status.Health {
	case graphql.DatadogAccountHealthOK:
		return "Services healthy, analyzing log patterns..."
	default:
		return "Processing..."
	}
}
