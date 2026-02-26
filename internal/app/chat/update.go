package chat

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/chat/msgs"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/tea/keymap"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle messages this model cares about.
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keymap.Tab) {
			cmds = append(cmds, m.toggleFocus())
			return tea.Batch(cmds...)
		}
		if m.focus == focusMessages {
			// Enter or esc returns to editor.
			if key.Matches(msg, keymap.Send) || key.Matches(msg, keymap.Exit) {
				cmds = append(cmds, m.setFocus(focusEditor))
				return tea.Batch(cmds...)
			}
			// Only forward to message list when it's focused.
			cmds = append(cmds, m.messageList.Update(msg))
			return tea.Batch(cmds...)
		}

	case emptyStatePollMsg:
		if !m.hasMessages() && m.db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			summary, err := m.db.DatadogAccountStatuses().GetSummary(ctx)
			if err == nil {
				m.policySummary = &summary
			}
			return m.pollEmptyState()
		}
		return nil // stop polling once messages exist

	case msgs.UserSubmittedInput:
		cmds = append(cmds, m.handleUserInput(msg))

	case conversationCreated:
		m.scope.Info("conversation created", "id", msg.conversationID)
		m.conversationID = msg.conversationID
		cmds = append(cmds, m.persistUserMessage(msg.input))

	case userMessagePersisted:
		m.scope.Debug("received userMessagePersisted", "message_id", msg.messageID)
		cmds = append(cmds, m.handlePersistedMessage(msg))

	case msgs.StreamCompleted:
		if m.session != nil && m.messageList.HasTurn(msg.TurnID) {
			m.session.RecordAssistantMessage(msg.Message)
		}

	case msgs.ToolResultMessagePersisted:
		if m.session == nil {
			m.session = chatclient.NewSession(m.conversationID, nil)
		}
		m.session.AppendMessage(msg.Message)

	case msgs.StreamFailed:
		return m.handleStreamFailed(msg)

	case tea.MouseClickMsg:
		// Click on the message list area focuses it.
		if m.hasMessages() && msg.Y >= m.originY && msg.Y < m.originY+m.height-m.inputBar.Height() {
			if m.focus != focusMessages {
				cmds = append(cmds, m.setFocus(focusMessages))
			}
		} else if msg.Y >= m.originY+m.height-m.inputBar.Height() {
			if m.focus != focusEditor {
				cmds = append(cmds, m.setFocus(focusEditor))
			}
		}
	}

	// Forward to children.
	cmds = append(cmds, m.inputBar.Update(msg))
	cmds = append(cmds, m.messageList.Update(msg))

	return tea.Batch(cmds...)
}

func (m *Model) handleStreamFailed(msg msgs.StreamFailed) tea.Cmd {
	errorClass := chatclient.ClassifyStreamError(msg.Err)
	m.scope.Warn("stream failed", "class", string(errorClass), "error", msg.Err)

	// Forward to round so it transitions to StateFailed.
	cmds := []tea.Cmd{m.messageList.Update(msg)}

	// Clean up orphaned messages from DB (async to avoid blocking UI).
	if last := m.messageList.LastRound(); last != nil {
		ids := last.LastTurnMessageIDs()
		if len(ids) > 0 {
			if m.session != nil {
				m.session.RemoveMessagesByID(ids)
			}
			db := m.db
			scope := m.scope
			cmds = append(cmds, func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
				defer cancel()
				for _, id := range ids {
					if err := db.Messages().Delete(ctx, id); err != nil {
						scope.Error("failed to delete orphaned message", "id", id, "error", err)
					}
				}
				return nil
			})
		}

		// Turn 1: remove round entirely, input bar restores text via pendingText.
		if !last.HasAssistantContent() {
			m.messageList.RemoveLastRound()
		}
		// Turn 2+: round stays visible with red error divider.
	}

	cmds = append(cmds, m.inputBar.Update(msg))
	cmds = append(cmds, appmsg.ErrorCmd(chatclient.UserFacingStreamError(msg.Err), msg.Err, false))
	return tea.Batch(cmds...)
}

// toggleFocus switches focus between editor and messages.
func (m *Model) toggleFocus() tea.Cmd {
	switch m.focus {
	case focusEditor:
		if !m.hasMessages() {
			return nil // nothing to focus
		}
		return m.setFocus(focusMessages)
	default:
		return m.setFocus(focusEditor)
	}
}

// setFocus sets focus to the given target.
func (m *Model) setFocus(f focus) tea.Cmd {
	m.focus = f
	switch f {
	case focusEditor:
		m.messageList.SetFocused(false)
		return m.inputBar.Focus()
	case focusMessages:
		m.inputBar.Blur()
		m.messageList.SetFocused(true)
		return nil
	}
	return nil
}
