package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domain"
)

// startSession scopes the API services to the account and opens a session
// context that is cancelled on shutdown. There is no local database or sync
// engine: reads and writes go straight to the control plane over GraphQL.
func (m *Model) startSession(accountID string) {
	if m.sessionCancel != nil {
		m.shutdown()
	}

	sessionCtx, cancel := context.WithCancel(m.ctx)
	m.sessionCtx = sessionCtx
	m.sessionCancel = cancel

	// Scope API services to the active account.
	m.services = m.services.WithAccountID(domain.AccountID(accountID))
	m.scope.Info("session started", "account_id", accountID)
}

// ensureRuntime scopes the session to the account and starts the status
// surfaces, which read from the account-scoped control-plane services.
func (m *Model) ensureRuntime(accountID string) (tea.Cmd, error) {
	m.startSession(accountID)
	return m.statusBar.SetServices(m.services), nil
}
