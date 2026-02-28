package chat

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/chat/msgs"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/tea/keymap"
)

type updateDispatch struct {
	cmd     tea.Cmd
	handled bool
	stop    bool
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) updateDispatch {
	if key.Matches(msg, keymap.Tab) {
		return updateDispatch{cmd: m.toggleFocus(), handled: true, stop: true}
	}

	if m.focus == focusMessages {
		// Enter or esc returns to editor.
		if key.Matches(msg, keymap.Send) || key.Matches(msg, keymap.Exit) {
			return updateDispatch{cmd: m.setFocus(focusEditor), handled: true, stop: true}
		}
		// Only forward to message list when it's focused.
		return updateDispatch{cmd: m.messageList.Update(msg), handled: true, stop: true}
	}

	return updateDispatch{}
}

func (m *Model) handleEmptyStatePoll() tea.Cmd {
	if m.hasMessages() || m.db == nil {
		return nil // stop polling once messages exist
	}
	return tea.Batch(m.fetchEmptyStateSummary(), m.pollEmptyState())
}

func (m *Model) fetchEmptyStateSummary() tea.Cmd {
	db := m.db
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
		return emptyStateSummaryMsg{summary: summary, err: err}
	}
}

func (m *Model) handleEmptyStateSummary(msg emptyStateSummaryMsg) {
	if msg.err != nil {
		return
	}
	m.policySummary = &msg.summary
}

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) tea.Cmd {
	// Click on the message list area focuses it.
	if m.hasMessages() && msg.Y >= m.originY && msg.Y < m.originY+m.height-m.inputBar.Height() {
		if m.focus != focusMessages {
			return m.setFocus(focusMessages)
		}
		return nil
	}
	if msg.Y >= m.originY+m.height-m.inputBar.Height() && m.focus != focusEditor {
		return m.setFocus(focusEditor)
	}
	return nil
}

func (m *Model) handleLifecycleMessage(msg tea.Msg) updateDispatch {
	switch msg := msg.(type) {
	case msgs.UserSubmittedInput:
		return updateDispatch{cmd: m.handleUserInput(msg), handled: true}

	case conversationCreated:
		m.scope.Info("conversation created", "id", msg.conversationID)
		m.conversationID = msg.conversationID
		return updateDispatch{cmd: m.persistUserMessage(msg.input), handled: true}

	case userMessagePersisted:
		m.scope.Debug("received userMessagePersisted", "message_id", msg.messageID)
		return updateDispatch{cmd: m.handlePersistedMessage(msg), handled: true}

	case msgs.StreamCompleted:
		if m.session != nil && m.messageList.HasTurn(msg.TurnID) {
			m.session.RecordAssistantMessage(msg.Message)
		}
		return updateDispatch{handled: true}

	case msgs.ToolResultMessagePersisted:
		if m.session == nil {
			m.session = corechat.NewSession(m.conversationID, nil)
		}
		m.session.AppendMessage(msg.Message)
		return updateDispatch{handled: true}

	case msgs.StreamFailed:
		return updateDispatch{cmd: m.handleStreamFailed(msg), handled: true, stop: true}
	}

	return updateDispatch{}
}
