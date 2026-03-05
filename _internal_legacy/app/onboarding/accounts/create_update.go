package accounts

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *CreateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case accountCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.scope.Error("failed to create account", "error", msg.err)
			m.err = msg.err
			return appevents.PublishErrorToastCmd("Failed to create account", msg.err, false)
		}
		_ = m.prefs.SetDefaultAccountID(msg.account.ID)
		org := m.org
		acc := msg.account
		m.scope.Info("account created", "id", acc.ID, "name", acc.Name)
		return func() tea.Msg { return bootstrap.AccountCreated{Org: org, Account: acc} }

	case tea.KeyPressMsg:
		if name, ok := stepkit.ParseCreateSubmit(msg, m.creating, m.input.Value()); ok {
			m.creating = true
			m.err = nil
			m.scope.Info("creating account", "name", name)
			return m.createAccount(name)
		}
	}

	return m.input.Update(msg)
}
