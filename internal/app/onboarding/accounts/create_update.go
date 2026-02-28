package accounts

import (
	tea "charm.land/bubbletea/v2"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

// Update handles messages.
func (m *CreateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case accountCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.scope.Error("failed to create account", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to create account", msg.err, false)
		}
		_ = m.prefs.SetDefaultAccountID(msg.account.ID)
		org := m.org
		acc := msg.account
		m.scope.Info("account created", "id", acc.ID, "name", acc.Name)
		return func() tea.Msg { return msgs.AccountCreated{Org: org, Account: acc} }

	case tea.KeyPressMsg:
		if m.creating {
			return nil
		}
		if msg.String() == "enter" {
			name := m.input.Value()
			if name == "" {
				return nil
			}
			m.creating = true
			m.err = nil
			m.scope.Info("creating account", "name", name)
			return m.createAccount(name)
		}
	}

	return m.input.Update(msg)
}
