package chat

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domain"
)

// SetSize updates the dimensions. This is a flexible component.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateLayout()
}

// SetOrigin sets the terminal-absolute position of this component's top-left corner.
func (m *Model) SetOrigin(x, y int) {
	m.originX = x
	m.originY = y
	m.updateLayout()
}

// updateLayout calculates sizes for children based on current dimensions.
func (m *Model) updateLayout() {
	// Input bar is fixed height.
	m.inputBar.SetWidth(m.width)
	inputBarHeight := m.inputBar.Height()

	// MessageList is flexible - gets remaining space (minus 1 for spacer between list and input bar).
	spacer := 0
	if m.hasMessages() {
		spacer = 1
	}
	messageListHeight := m.height - inputBarHeight - spacer
	if messageListHeight < 0 {
		messageListHeight = 0
	}
	m.messageList.SetSize(m.width, messageListHeight)
	m.messageList.SetOrigin(m.originX, m.originY)
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.focus == focusMessages {
		return []key.Binding{scrollUp, focusInputBar}
	}
	if m.hasMessages() {
		return append(m.inputBar.ShortHelp(), focusChat)
	}
	return m.inputBar.ShortHelp()
}

// ConversationID returns the current conversation ID.
func (m *Model) ConversationID() domain.ConversationID {
	return m.conversationID
}

// CancelActiveRound cancels the active round if one exists.
// Returns true if a round was cancelled and a command for async cleanup.
func (m *Model) CancelActiveRound() (bool, tea.Cmd) {
	if !m.messageList.HasActiveRound() {
		return false, nil
	}

	last := m.messageList.LastRound()
	m.messageList.CancelActiveRound()

	var cleanupCmd tea.Cmd
	if last != nil {
		ids := last.LastTurnMessageIDs()
		if m.session != nil {
			m.session.RemoveMessagesByID(ids)
		}
		cleanupCmd = m.cleanupOrphanedMessages(ids)

		// Turn 1: remove round entirely (no assistant content to show).
		if !last.HasAssistantContent() {
			m.messageList.RemoveLastRound()
		}
	}

	return true, cleanupCmd
}

// hasMessages returns true if there are messages to display.
func (m *Model) hasMessages() bool {
	return m.messageList.Len() > 0
}

type orphanedMessagesCleanupCompleted struct {
	ids []domain.MessageID
	err error
}

func (m *Model) cleanupOrphanedMessages(ids []domain.MessageID) tea.Cmd {
	if len(ids) == 0 || m.runtimeDeps.OrphanCleaner == nil {
		return nil
	}
	cleaner := m.runtimeDeps.OrphanCleaner
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
		defer cancel()
		return orphanedMessagesCleanupCompleted{
			ids: ids,
			err: cleaner.CleanupMessages(ctx, ids),
		}
	}
}
