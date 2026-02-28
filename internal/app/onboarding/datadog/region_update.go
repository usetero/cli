package datadog

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
)

// Update handles messages.
func (m *RegionModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, regionUpKey):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, regionDownKey):
			if m.selected < len(domain.DatadogRegions)-1 {
				m.selected++
			}
		case key.Matches(msg, regionSelectKey):
			site := domain.DatadogRegions[m.selected].Site
			m.scope.Info("datadog region selected", "site", site)
			return func() tea.Msg {
				return msgs.DatadogRegionSelected{Site: site}
			}
		}
	}
	return nil
}
