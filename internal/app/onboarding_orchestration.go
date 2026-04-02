package app

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

func (m *Model) handleOnboardingMessage(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case bootstrap.OrgSelected:
		return tea.Batch(m.statusBar.Update(msg), m.activateOrg(msg.Org.ID, msg)), true

	case bootstrap.OrgCreated:
		return tea.Batch(m.statusBar.Update(msg), m.activateOrg(msg.Org.ID, msg)), true

	case bootstrap.AccountSelected:
		// Forward to onboarding orchestrator; runtime init happens at EnsureRuntime gate.
		if m.onboarding != nil {
			return m.onboarding.Update(msg), true
		}
		return nil, true

	case bootstrap.EnsureRuntime:
		m.scope.Info("ensuring runtime", "account_id", msg.Account.ID.String())
		start := time.Now()
		catalogCmd, err := m.ensureRuntime(msg.Account.ID.String())
		if err != nil {
			m.scope.Error("failed to ensure runtime", "error", err)
			return appevents.PublishErrorToastCmd("Failed to initialize account runtime", err, true), true
		}
		m.scope.Info("runtime ensured", "account_id", msg.Account.ID.String(), "elapsed_ms", time.Since(start).Milliseconds())
		if m.onboarding != nil {
			return tea.Batch(
				catalogCmd,
				func() tea.Msg { return bootstrap.RuntimeReady(msg) },
			), true
		}
		return catalogCmd, true

	case bootstrap.OnboardingComplete:
		m.state = stateChat
		m.user = msg.User
		m.account = msg.Account
		m.scope.Info("onboarding complete",
			"org", msg.Org.Name,
			"account", msg.Account.Name,
		)

		// Create chat model (sizing happens via updateLayout)
		m.chat = m.newChat()

		// Size the new chat component
		m.updateLayout()

		return m.chat.Init(), true
	}

	return nil, false
}

func (m *Model) handleStreamCompleted(msg tea.Msg) {
	stream, ok := msg.(msgs.StreamCompleted)
	if !ok {
		return
	}

	if stream.Title != "" && m.db != nil && m.chat != nil {
		m.statusBar.SetTitle(stream.Title)
		m.windowTitle = "Tero: " + stream.Title
		db := m.db
		conversationID := m.chat.ConversationID()
		title := stream.Title
		scope := m.scope
		ctx := m.ctx
		// Persist title in background using immutable captured values.
		go func() {
			writeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := db.Conversations().UpdateTitle(writeCtx, conversationID, title); err != nil {
				scope.Error("failed to update conversation title", "error", err)
			}
		}()
	}
	// Update context window usage in statusbar
	if stream.InputTokens > 0 && stream.ContextWindow > 0 {
		pct := (stream.InputTokens*100 + stream.ContextWindow - 1) / stream.ContextWindow // round up
		m.statusBar.SetContextPercent(pct)
	}
}
