package chat

import (
	"context"

	"charm.land/bubbles/v2/key"
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
// Returns true if a round was cancelled.
func (m *Model) CancelActiveRound() bool {
	if !m.messageList.HasActiveRound() {
		return false
	}

	last := m.messageList.LastRound()
	m.messageList.CancelActiveRound()

	// Clean up orphaned messages from DB — same as StreamFailed handler.
	if last != nil {
		ids := last.LastTurnMessageIDs()
		if m.session != nil {
			m.session.RemoveMessagesByID(ids)
		}
		ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
		defer cancel()
		for _, id := range ids {
			if err := m.db.Messages().Delete(ctx, id); err != nil {
				m.scope.Error("failed to delete orphaned message", "id", id, "error", err)
			}
		}

		// Turn 1: remove round entirely (no assistant content to show).
		if !last.HasAssistantContent() {
			m.messageList.RemoveLastRound()
		}
	}

	return true
}

// hasMessages returns true if there are messages to display.
func (m *Model) hasMessages() bool {
	return m.messageList.Len() > 0
}
