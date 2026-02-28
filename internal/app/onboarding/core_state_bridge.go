package onboarding

import (
	"github.com/usetero/cli/internal/domain"
)

func (m *Model) toCoreState() onboardingState { return m.state }

func (m *Model) applyCoreState(state onboardingState) {
	m.state = state
	m.syncServicesToState()
}

func (m *Model) syncServicesToState() {
	accountID := domain.AccountID("")
	if m.state.Account != nil {
		accountID = m.state.Account.ID
	}
	m.services = m.services.WithAccountID(accountID)
}
