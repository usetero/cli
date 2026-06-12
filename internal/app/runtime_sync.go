package app

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/usecase"
	chattools "github.com/usetero/cli/internal/app/chattools"
	chatboundary "github.com/usetero/cli/internal/boundary/chat"
	"github.com/usetero/cli/internal/domain"
)

// startSync starts the syncer with the open database.
func (m *Model) startSync(accountID string) error {
	if m.db == nil {
		return fmt.Errorf("database not open")
	}
	if m.sessionCancel != nil {
		m.shutdown()
		if err := m.openDatabase(accountID); err != nil {
			return err
		}
	}

	// Create a session context that is cancelled on shutdown.
	sessionCtx, cancel := context.WithCancel(m.ctx)
	m.sessionCtx = sessionCtx
	m.sessionCancel = cancel

	if err := m.syncer.Start(sessionCtx, m.db, accountID, nil); err != nil {
		cancel()
		m.sessionCtx = nil
		m.sessionCancel = nil
		return err
	}
	m.scope.Info("syncer started", "account_id", accountID)

	// Scope API services to the active account.
	m.services = m.services.WithAccountID(domain.AccountID(accountID))

	return nil
}

// ensureRuntime opens account database, starts sync, and initializes dependent runtime services.
func (m *Model) ensureRuntime(accountID string) (tea.Cmd, error) {
	if err := m.openDatabase(accountID); err != nil {
		return nil, err
	}

	if err := m.startSync(accountID); err != nil {
		return nil, err
	}

	// Start catalog status polling now that db is ready
	catalogCmd := m.statusBar.SetDB(m.db)

	// Create tool registry with executors
	m.toolRegistry = chattools.NewRegistry(
		chattools.NewQueryTool(m.db, m.scope),
		chattools.NewShowTool(m.db),
		map[string]chattools.ActionTool{
			"set_service_enabled": chattools.NewSetServiceEnabledAction(m.db),
			"approve_policy": chattools.NewApprovePolicyAction(m.db, func() string {
				if m.user != nil {
					return m.user.ID
				}
				return ""
			}),
		},
	)

	// Create chat client with tool definitions
	m.chatClient = chatboundary.NewClient(m.cfg.ChatEndpoint, m.authService, m.scope, m.toolRegistry.Definitions()).
		WithAccountID(domain.AccountID(accountID))
	m.runtimeDeps = usecase.NewRuntimeDeps(m.chatClient).WithEffectContext(m.sessionCtx)

	return catalogCmd, nil
}
