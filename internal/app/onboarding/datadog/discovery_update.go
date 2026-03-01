package datadog

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *DiscoveryModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case pollTickMsg:
		return m.pollStatus()

	case statusMsg:
		return m.handleStatus(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd

	case tea.KeyPressMsg:
		if msg.String() == "r" && m.err != nil {
			m.scope.Debug("retrying discovery")
			m.err = nil
			return m.Init()
		}
	}

	return m.progress.Update(msg)
}

func (m *DiscoveryModel) handleStatus(msg statusMsg) tea.Cmd {
	if msg.err != nil {
		m.scope.Error("discovery status check failed", "error", msg.err)
		m.err = msg.err
		return appevents.ErrorCmd("Failed to check discovery status", msg.err, false)
	}
	m.status = msg.status

	if msg.status.ReadyForUse {
		m.scope.Info("datadog discovery complete", "services", msg.status.ServiceCount)
		return func() tea.Msg { return bootstrap.DatadogDiscoveryComplete{} }
	}
	return m.schedulePoll()
}
