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

	// Start status polling: drawer tabs read from the account-scoped
	// control-plane services; the sync indicator still reads the runtime db.
	catalogCmd := tea.Batch(
		m.statusBar.SetServices(m.services),
		m.statusBar.SetDB(m.db),
	)

	// Create tool registry. All tools are GraphQL-backed: the read tools query
	// the control-plane catalog and set_service_enabled is a synchronous
	// mutation. Policy approval moved to the issue model and is no longer a
	// chat action.
	m.toolRegistry = chattools.NewRegistry(
		map[string]chattools.ActionTool{
			"list_services":       chattools.NewListServicesTool(m.services),
			"list_issues":         chattools.NewListIssuesTool(m.services),
			"list_checks":         chattools.NewListChecksTool(m.services),
			"list_edge_instances": chattools.NewListEdgeInstancesTool(m.services),
			"account_status":      chattools.NewAccountStatusTool(m.services),
			"set_service_enabled": chattools.NewSetServiceEnabledAction(m.services.Services),
		},
	)

	// Create chat client with tool definitions
	m.chatClient = chatboundary.NewClient(m.cfg.ChatEndpoint, m.authService, m.scope, m.toolRegistry.Definitions()).
		WithAccountID(domain.AccountID(accountID))
	m.runtimeDeps = usecase.NewRuntimeDeps(m.chatClient).WithEffectContext(m.sessionCtx)

	return catalogCmd, nil
}
