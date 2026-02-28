package onboarding

import (
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

func (m *Model) toCoreState() bootstrap.State {
	return bootstrap.State{
		User:      m.state.user,
		Org:       m.state.org,
		Account:   m.state.account,
		Workspace: m.state.workspace,
		DDSite:    m.state.ddSite,
		DDAPIKey:  m.state.ddAPIKey,
		DDAccount: m.state.ddAccount,
	}
}

func (m *Model) applyCoreState(state bootstrap.State) {
	m.state.user = state.User
	m.state.org = state.Org
	m.state.account = state.Account
	m.state.workspace = state.Workspace
	m.state.ddSite = state.DDSite
	m.state.ddAPIKey = state.DDAPIKey
	m.state.ddAccount = state.DDAccount
	m.syncServicesToState()
}

func (m *Model) syncServicesToState() {
	accountID := domain.AccountID("")
	if m.state.account != nil {
		accountID = m.state.account.ID
	}
	m.services = m.services.WithAccountID(accountID)
}
