package datadog

import (
	tea "charm.land/bubbletea/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

func (m *APIKeyModel) validateAPIKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		valid, errorMsg, err := m.services.DatadogAccounts.ValidateAPIKey(m.ctx, graphql.ValidateAPIKeyInput{
			APIKey: apiKey,
			Site:   m.site.String(),
		})
		if err != nil {
			return apiKeyValidatedMsg{err: err}
		}
		return apiKeyValidatedMsg{valid: valid, errorMsg: errorMsg}
	}
}
