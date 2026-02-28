package datadog

import (
	tea "charm.land/bubbletea/v2"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *AppKeyModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case accountCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.scope.Error("failed to create datadog account", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to create Datadog account", msg.err, false)
		}
		m.scope.Info("datadog account created", "id", msg.datadogAccountID)
		ddAccountID := msg.datadogAccountID
		return func() tea.Msg {
			return bootstrap.DatadogAccountCreated{DatadogAccountID: ddAccountID}
		}

	case tea.KeyPressMsg:
		if m.creating {
			return nil
		}
		if msg.String() == "enter" {
			appKey := m.input.Value()
			if appKey == "" {
				return nil
			}
			m.creating = true
			m.err = nil
			m.scope.Info("creating datadog account")
			return m.createAccount(appKey)
		}
	}

	return m.input.Update(msg)
}
