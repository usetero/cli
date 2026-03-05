package onboarding

import "github.com/usetero/cli/internal/domain"

func (m *Model) syncServicesToState() {
	accountID := domain.AccountID("")
	if m.state.Account != nil {
		accountID = m.state.Account.ID
	}
	m.services = m.services.WithAccountID(accountID)
}
