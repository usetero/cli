package round

import (
	"time"

	tea "charm.land/bubbletea/v2"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/domain"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	// Terminal states — no state transitions, no forwarding.
	if m.state == StateCancelled || m.state == StateFailed {
		return nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgs.StreamCompleted:
		next, changed := reduceOnStreamCompleted(m.state, m.isOurTurn(msg.TurnID), msg.Message.StopReason)
		if changed {
			if m.session != nil {
				m.session.RecordAssistantMessage(msg.Message)
			}
			m.state = next
			m.endTime = time.Now()
			m.scope.Info("round complete", "stop_reason", msg.Message.StopReason)
		} else if m.isOurTurn(msg.TurnID) && m.session != nil {
			// tool_use path still records assistant for next-turn history
			m.session.RecordAssistantMessage(msg.Message)
		}

	case msgs.StreamFailed:
		next, changed := reduceOnStreamFailed(m.state, m.isOurTurn(msg.TurnID))
		if changed {
			m.state = next
			m.lastErr = msg.Err
			m.endTime = time.Now()
			m.scope.Info("round failed", "error", msg.Err)
		}

	case msgs.ToolResultsReady:
		next, changed := reduceOnToolResultsReady(m.state, m.isOurTurn(msg.TurnID))
		if changed {
			m.state = next
			cmds = append(cmds, m.startNextTurn(msg.Results))
		}

	case nextTurnReady:
		next, changed := reduceOnNextTurnReady(m.state, msg.roundID == m.id)
		if changed {
			m.state = next
			cmds = append(cmds, m.handleNextTurnReady(msg))
		}
	}

	// Forward thinking ticks while active
	if m.IsActive() {
		cmds = append(cmds, m.thinking.Update(msg))
	}

	// Forward to all turns
	for _, t := range m.turns {
		cmds = append(cmds, t.Update(msg))
	}

	return tea.Batch(cmds...)
}

// isOurTurn checks if the given turn ID belongs to this round.
func (m *Model) isOurTurn(turnID domain.MessageID) bool {
	for _, t := range m.turns {
		if t.UserMessageID() == turnID {
			return true
		}
	}
	return false
}

// HasTurn reports whether this round owns turnID.
func (m *Model) HasTurn(turnID domain.MessageID) bool {
	return m.isOurTurn(turnID)
}
