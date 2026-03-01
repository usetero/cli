package datadog

import (
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"

	api "github.com/usetero/cli/internal/boundary/graphql"
)

func (m *AppKeyModel) createAccount(appKey string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		ddAccount, err := m.services.DatadogAccounts.CreateAccount(m.ctx, api.CreateDatadogAccountInput{
			ID:        id,
			AccountID: m.account.ID,
			Name:      m.account.Name,
			Site:      m.site.String(),
			APIKey:    m.apiKey,
			AppKey:    appKey,
		})
		if err != nil {
			return accountCreatedMsg{err: err}
		}
		return accountCreatedMsg{datadogAccountID: ddAccount.ID}
	}
}
