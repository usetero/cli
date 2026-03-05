package accounts

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

func (m *SelectModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accounts, err := m.services.Accounts.List(m.ctx, m.org.ID)
		if err != nil {
			return remotelist.LoadResult{Err: err}
		}

		items := make([]remotelist.Item, len(accounts))
		for i, acc := range accounts {
			items[i] = acc
		}
		return remotelist.LoadResult{Items: items}
	}
}

func (m *SelectModel) emitSelected(acc domain.Account) tea.Cmd {
	org := m.org
	m.scope.Info("account selected", "id", acc.ID, "name", acc.Name)
	return func() tea.Msg {
		return bootstrap.AccountSelected{Org: org, Account: acc}
	}
}
