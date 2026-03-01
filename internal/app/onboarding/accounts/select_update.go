package accounts

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// Update handles messages.
func (m *SelectModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case remotelist.LoadResult:
		return m.handleLoadResult(msg)
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m.list.Update(msg)
}

func (m *SelectModel) handleLoadResult(msg remotelist.LoadResult) tea.Cmd {
	if msg.Err != nil {
		m.scope.Error("failed to load accounts", "error", msg.Err)
		return tea.Batch(m.list.Update(msg), appevents.ErrorCmd("Failed to load accounts", msg.Err, false))
	}

	m.accounts = stepkit.CastItems[domain.Account](msg.Items)
	m.scope.Info("accounts loaded", "count", len(m.accounts))

	if len(m.accounts) == 0 {
		m.scope.Debug("no accounts found")
		org := m.org
		return func() tea.Msg { return bootstrap.NoAccounts{Org: org} }
	}

	if prefAccount := findAccountByID(m.accounts, m.prefs.GetDefaultAccountID()); prefAccount != nil {
		m.scope.Debug("using saved account preference", "id", prefAccount.ID)
		return m.emitSelected(*prefAccount)
	}

	if len(m.accounts) == 1 {
		m.scope.Debug("auto-selected account (only one)")
		acc := m.accounts[0]
		_ = m.prefs.SetDefaultAccountID(acc.ID)
		return m.emitSelected(acc)
	}

	return m.list.Update(msg)
}

func (m *SelectModel) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.list.IsLoading() {
		return nil
	}

	switch msg.String() {
	case "enter":
		if item := m.list.SelectedItem(); item != nil {
			if acc, ok := item.(domain.Account); ok {
				_ = m.prefs.SetDefaultAccountID(acc.ID)
				return m.emitSelected(acc)
			}
		}
	case "n":
		org := m.org
		return func() tea.Msg { return bootstrap.NoAccounts{Org: org} }
	case "r":
		if m.list.HasError() {
			m.scope.Debug("retrying account load")
			return m.list.Retry()
		}
	}

	return nil
}
