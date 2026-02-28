package accounts

import (
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"

	"github.com/usetero/cli/internal/api"
)

func (m *CreateModel) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		account, err := m.services.Accounts.Create(m.ctx, api.CreateAccountInput{
			ID:             id,
			OrganizationID: m.org.ID,
			Name:           name,
		})
		if err != nil {
			return accountCreatedMsg{err: err}
		}
		return accountCreatedMsg{account: *account}
	}
}
