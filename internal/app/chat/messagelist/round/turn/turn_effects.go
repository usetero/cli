package turn

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/usecase"
	"github.com/usetero/cli/internal/domain"
)

// persistAssistantMessage saves the assistant message to the database.
func (m *Model) persistAssistantMessage(msg *domain.Message) tea.Cmd {
	if msg == nil {
		return nil
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
		defer cancel()

		msgID, err := m.assistantPersister.PersistAssistant(ctx, usecase.PersistAssistantInput{
			AccountID:      m.accountID,
			ConversationID: m.conversationID,
			Message:        *msg,
		})
		if err != nil {
			m.scope.Error("failed to persist assistant message", "error", err)
			return nil
		}

		m.scope.Info("assistant message persisted", "message_id", msgID)
		return assistantPersisted{
			turnID:    m.userMessage.ID(),
			messageID: msgID,
		}
	}
}

// fireToolResults fires ToolResultsReady for the round to handle.
// Guarded by firedToolResults to ensure it can only fire once per turn.
func (m *Model) fireToolResults() tea.Cmd {
	return m.toolTracker.fire(m.userMessage.ID(), func(message string) {
		m.scope.Warn(message)
	})
}

// collectToolUseIDs returns the count and set of tool_use IDs in content.
func collectToolUseIDs(content []domain.Block) (int, map[string]bool) {
	ids := make(map[string]bool)
	for _, b := range content {
		if b.Type == domain.BlockTypeToolUse && b.ToolUse != nil && b.ToolUse.ID != "" {
			ids[b.ToolUse.ID] = true
		}
	}
	return len(ids), ids
}

func (m *Model) reportProtocolViolation(reason string, kv ...any) {
	m.protocolViolationCount++
	fields := []any{
		"reason", reason,
		"count", m.protocolViolationCount,
	}
	fields = append(fields, kv...)
	m.scope.Warn("protocol violation", fields...)
}
