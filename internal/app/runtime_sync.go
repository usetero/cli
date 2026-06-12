package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/usecase"
	chattools "github.com/usetero/cli/internal/app/chattools"
	chatboundary "github.com/usetero/cli/internal/boundary/chat"
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

// ensureRuntime scopes the session to the account and initializes dependent
// runtime services (status surfaces, chat tools, chat client).
func (m *Model) ensureRuntime(accountID string) (tea.Cmd, error) {
	m.startSession(accountID)

	// Drawer tabs read from the account-scoped control-plane services.
	catalogCmd := m.statusBar.SetServices(m.services)

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
