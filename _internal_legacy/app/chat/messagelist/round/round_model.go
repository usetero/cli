package round

import (
	"time"

	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks"
	"github.com/usetero/cli/internal/domain"
)

// Blocks returns all visual blocks from all turns in this round.
// The thinking animation is appended at the end while the round is active.
func (m *Model) Blocks() []block.Block {
	var result []block.Block
	for _, t := range m.turns {
		result = append(result, t.Blocks()...)
	}
	if m.IsActive() {
		result = append(result, blocks.NewThinkingAnimBlock(m.thinking))
	}
	return result
}

// SetWidth sets the width for all turns.
func (m *Model) SetWidth(width int) {
	m.width = width
	for _, t := range m.turns {
		t.SetWidth(width)
	}
}

// Cancel stops all in-flight turns and marks the round cancelled.
func (m *Model) Cancel() {
	for _, t := range m.turns {
		t.Cancel()
	}
	m.state = StateCancelled
	m.endTime = time.Now()
	m.scope.Info("round cancelled")
}

// State returns the round's current state.
func (m *Model) State() State {
	return m.state
}

// ID returns the round's ID (first user message ID).
func (m *Model) ID() domain.MessageID {
	return m.id
}

// Err returns the error that caused the round to fail, or nil.
func (m *Model) Err() error {
	return m.lastErr
}

// HasAssistantContent returns true if any turn has assistant blocks.
func (m *Model) HasAssistantContent() bool {
	for _, t := range m.turns {
		if len(t.Blocks()) > 1 { // more than just the user message block
			return true
		}
	}
	return false
}

// LastTurnMessageIDs returns the message IDs that should be deleted on failure.
// For turn 1: the user message ID.
// For turn 2+: the tool result message ID (current turn) + the previous turn's assistant message ID.
func (m *Model) LastTurnMessageIDs() []domain.MessageID {
	if len(m.turns) == 0 {
		return nil
	}

	last := m.turns[len(m.turns)-1]

	if len(m.turns) == 1 {
		// Turn 1: just the user message
		return []domain.MessageID{last.UserMessageID()}
	}

	// Turn 2+: tool result message + previous assistant message
	prev := m.turns[len(m.turns)-2]
	ids := []domain.MessageID{last.UserMessageID()}
	if aid := prev.AssistantMessageID(); aid != "" {
		ids = append(ids, aid)
	}
	return ids
}

// Duration returns the elapsed time for this round.
func (m *Model) Duration() time.Duration {
	if m.endTime.IsZero() {
		return time.Since(m.startTime)
	}
	return m.endTime.Sub(m.startTime)
}
