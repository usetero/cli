package datadog

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

func (m *DiscoveryModel) pollStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.services.DatadogAccounts.GetStatus(m.ctx, m.datadogAccountID)
		return discoveryStatusLoadedMsg{status: status, err: err}
	}
}

func (m *DiscoveryModel) schedulePoll() tea.Cmd {
	return tea.Tick(discoveryPollInterval, func(time.Time) tea.Msg {
		return discoveryPollTickMsg{}
	})
}
