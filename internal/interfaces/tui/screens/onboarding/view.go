package onboarding

import (
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/chrome"
)

func (m *Model) View() tea.View {
	view := m.baseView()
	if m.loadErr != nil && m.route != routeError {
		view.Content = lipgloss.JoinVertical(
			lipgloss.Left,
			view.Content,
			"",
			m.theme.Text.Error.Render("Error: "+m.loadErr.Error()),
		)
	}
	return view
}

func (m *Model) baseView() tea.View {
	switch m.route {
	case routeLoading:
		return tea.NewView(m.theme.Text.Muted.Render("Loading onboarding state..."))
	case routeRole:
		return m.role.View()
	case routeTenancy:
		return m.tenancy.View()
	case routeIntegrations:
		return m.integrations.View()
	case routePowerSyncReady:
		return m.powersync.View()
	case routeDone:
		return tea.NewView(m.theme.Text.Section.Render("Welcome to Tero"))
	case routePlaceholder:
		return tea.NewView(lipgloss.JoinVertical(
			lipgloss.Left,
			m.theme.Text.Section.Render("Onboarding step: "+string(m.step)),
			"",
			m.theme.Text.Muted.Render("Not implemented yet."),
		))
	case routeError:
		return tea.NewView(chrome.RenderErrorCard(
			m.theme,
			"Failed to load onboarding state.",
			m.loadErr.Error(),
		))
	default:
		return tea.NewView(m.theme.Text.Error.Render("Unknown onboarding route"))
	}
}
