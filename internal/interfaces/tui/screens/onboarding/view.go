package onboarding

import (
	"charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/chrome"
	"github.com/usetero/cli/internal/interfaces/tui/present"
)

func (m *Model) Layout() chrome.BodyLayout {
	align := chrome.AlignBottom
	if m.route == routeError {
		align = chrome.AlignCenter
	}
	return chrome.BodyLayout{
		WidthMode:     chrome.WidthIntrinsic,
		HeightMode:    chrome.HeightIntrinsic,
		VerticalAlign: align,
		MaxWidth:      80,
	}
}

func (m *Model) View() tea.View {
	view := m.baseView()
	if m.loadErr != nil && m.route != routeError {
		view = present.View(m.theme, present.StackGap(
			1,
			present.Raw(view.Content),
			present.Error("Error: "+m.loadErr.Error()),
		))
	}
	return view
}

func (m *Model) baseView() tea.View {
	switch m.route {
	case routeLoading:
		return m.loading.View()
	case routeRole:
		return m.role.View()
	case routeTenancy:
		return m.tenancy.View()
	case routeIntegrations:
		return m.integrations.View()
	case routePowerSyncReady:
		return m.powersync.View()
	case routeDone:
		return present.View(m.theme, present.Notice("Welcome to Tero", "Onboarding is complete."))
	case routePlaceholder:
		return present.View(m.theme, present.Notice("Onboarding step: "+string(m.step), "Not implemented yet."))
	case routeError:
		return present.View(m.theme, present.ErrorCard(present.BlockGap(
			1,
			present.Error("Failed to load onboarding state."),
			present.Body(m.loadErr.Error()),
		)))
	default:
		return tea.NewView(m.theme.Text.Error.Render("Unknown onboarding route"))
	}
}
