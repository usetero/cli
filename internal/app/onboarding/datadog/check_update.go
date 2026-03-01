package datadog

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *CheckModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case datadogCheckResultMsg:
		return m.handleCheckResult(msg)

	case tea.KeyPressMsg:
		if msg.String() == "r" && m.err != nil {
			m.scope.Debug("retrying datadog check")
			m.err = nil
			return m.checkDatadog()
		}
	}
	return nil
}

func (m *CheckModel) handleCheckResult(msg datadogCheckResultMsg) tea.Cmd {
	if msg.err != nil {
		m.scope.Error("datadog check failed", "error", msg.err)
		m.err = msg.err
		return appevents.PublishErrorToastCmd("Failed to check Datadog status", msg.err, false)
	}
	if msg.hasDatadog {
		m.scope.Info("datadog configured")
		return func() tea.Msg { return bootstrap.DatadogReady{} }
	}
	m.scope.Info("datadog setup required")
	return func() tea.Msg { return bootstrap.DatadogNeeded{} }
}
