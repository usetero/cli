package datadog

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *APIKeyModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case apiKeyValidatedMsg:
		m.validating = false
		if msg.err != nil {
			m.scope.Error("api key validation failed", "error", msg.err)
			m.err = msg.err
			return appevents.PublishErrorToastCmd("Failed to validate API key", msg.err, false)
		}
		if !msg.valid {
			m.scope.Info("api key invalid", "reason", msg.errorMsg)
			m.err = &validationError{msg.errorMsg}
			return nil
		}
		m.scope.Info("api key validated")
		apiKey := m.input.Value()
		return func() tea.Msg {
			return bootstrap.DatadogAPIKeyEntered{APIKey: apiKey}
		}

	case tea.KeyPressMsg:
		if m.validating {
			return nil
		}
		if msg.String() == "enter" {
			apiKey := m.input.Value()
			if apiKey == "" {
				return nil
			}
			m.validating = true
			m.err = nil
			m.scope.Info("validating api key")
			return m.validateAPIKey(apiKey)
		}
	}

	return m.input.Update(msg)
}
