package datadog

import tea "charm.land/bubbletea/v2"

func (m *CheckModel) checkDatadog() tea.Cmd {
	return func() tea.Msg {
		hasDatadog, err := m.services.DatadogAccounts.HasAccount(m.ctx, m.account.ID)
		if err != nil {
			return datadogCheckCompletedMsg{err: err}
		}
		return datadogCheckCompletedMsg{hasDatadog: hasDatadog}
	}
}
