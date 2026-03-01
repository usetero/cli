package app

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/usecase"
	chattools "github.com/usetero/cli/internal/app/chattools"
	chatboundary "github.com/usetero/cli/internal/boundary/chat"
	psapi "github.com/usetero/cli/internal/boundary/powersync"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/upload"
)

// startSync starts the syncer and uploader with the open database.
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
	m.sessionCancel = cancel

	if err := m.syncer.Start(sessionCtx, m.db, accountID, nil); err != nil {
		cancel()
		m.sessionCancel = nil
		return err
	}
	m.scope.Info("syncer started", "account_id", accountID)

	// Scope API services to the active account.
	m.services = m.services.WithAccountID(domain.AccountID(accountID))

	// Create PowerSync API client for write checkpoints
	psClient := psapi.NewClient(m.cfg.PowerSyncEndpoint)

	// Create and start uploader
	syncer := m.syncer
	scope := m.scope
	uploader := upload.New(
		m.db,
		psClient,
		m.authService,
		m.services.Conversations,
		m.services.Messages,
		m.services.Services,
		m.services.Policies,
		scope,
		upload.WithBatchCompletedHook(func(ctx context.Context) error {
			return syncer.NotifyUploadCompleted(ctx)
		}),
	)
	m.uploader = uploader
	go func() {
		if err := uploader.Run(sessionCtx); err != nil && !errors.Is(err, context.Canceled) {
			scope.Error("uploader error", "error", err)
		}
	}()
	scope.Info("uploader started", "account_id", accountID)

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
	m.runtimeDeps = usecase.NewRuntimeDeps(m.db, m.chatClient)

	return catalogCmd, nil
}
