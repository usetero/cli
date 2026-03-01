package chat

import (
	"context"

	tea "charm.land/bubbletea/v2"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/usecase"
	appevents "github.com/usetero/cli/internal/app/events"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		res := m.handleKeyPress(msg)
		if res.stop {
			return res.cmd
		}
		if res.handled && res.cmd != nil {
			cmds = append(cmds, res.cmd)
		}

	case emptyStatePollTickMsg:
		return m.handleEmptyStatePoll()

	case emptyStateSummaryLoadedMsg:
		m.handleEmptyStateSummary(msg)
		return nil

	case tea.MouseClickMsg:
		if cmd := m.handleMouseClick(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	default:
		res := m.handleLifecycleMessage(msg)
		if res.stop {
			return res.cmd
		}
		if res.handled && res.cmd != nil {
			cmds = append(cmds, res.cmd)
		}
	}

	// Forward to children.
	cmds = append(cmds, m.inputBar.Update(msg))
	cmds = append(cmds, m.messageList.Update(msg))

	return tea.Batch(cmds...)
}

func (m *Model) handleStreamFailed(msg msgs.StreamFailed) tea.Cmd {
	m.scope.Warn("stream failed", "class", usecase.ClassifyStreamError(m.runtimeDeps.StreamErrorMapper, msg.Err), "error", msg.Err)

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
	cmds = append(cmds, appevents.PublishErrorToastCmd(usecase.UserFacingStreamError(m.runtimeDeps.StreamErrorMapper, msg.Err), msg.Err, false))
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
