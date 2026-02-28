package round

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn"
	"github.com/usetero/cli/internal/app/chat/msgs"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
)

// startNextTurn persists tool results and creates the next turn using in-memory history.
func (m *Model) startNextTurn(results []domaintools.Result) tea.Cmd {
	m.scope.Info("starting next turn", "result_count", len(results))
	for _, summary := range summarizeToolResults(results) {
		m.scope.Debug("next turn tool result", "summary", summary)
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
		defer cancel()

		// Convert to domain format and persist
		domainResults := make([]domain.ToolResult, len(results))
		for i, r := range results {
			domainResults[i] = domain.ToolResult{
				ToolUseID: r.ToolUseID,
				IsError:   r.IsError(),
				Content:   r.ToMap(),
			}
			if r.Error != nil {
				domainResults[i].Error = r.Error.Message
			}
		}

		msgID, err := m.db.Messages().CreateToolResultMessage(ctx, m.accountID, m.conversationID, domainResults)
		if err != nil {
			// Durability failure should not block the active chat loop.
			m.scope.Error("failed to create tool result message", "error", err)
			msgID = domain.NewMessageID()
		}

		if m.session == nil {
			m.session = corechat.NewSession(m.conversationID, nil)
		}
		toolResultMessage := m.session.AppendUserToolResultsMessage(msgID, domainResults)
		messages := m.session.Messages()
		for _, summary := range summarizeHistory(messages) {
			m.scope.Debug("next turn history", "summary", summary)
		}

		return nextTurnReady{
			roundID:           m.id,
			messageID:         msgID,
			results:           results,
			messages:          messages,
			toolResultMessage: toolResultMessage,
		}
	}
}

func summarizeToolResults(results []domaintools.Result) []string {
	out := make([]string, 0, len(results))
	for _, r := range results {
		rows := -1
		if rawRows, ok := r.Content["rows"]; ok {
			if list, ok := rawRows.([]map[string]any); ok {
				rows = len(list)
			} else if listAny, ok := rawRows.([]any); ok {
				rows = len(listAny)
			}
		}
		if rows >= 0 {
			out = append(out, fmt.Sprintf("tool_use_id=%s is_error=%t rows=%d", r.ToolUseID, r.IsError(), rows))
			continue
		}
		out = append(out, fmt.Sprintf("tool_use_id=%s is_error=%t", r.ToolUseID, r.IsError()))
	}
	return out
}

func summarizeHistory(messages []domain.Message) []string {
	out := make([]string, 0, len(messages))
	for _, msg := range messages {
		blockKinds := make([]string, 0, len(msg.Content))
		for _, b := range msg.Content {
			blockKinds = append(blockKinds, string(b.Type))
		}
		out = append(out, fmt.Sprintf(
			"id=%s role=%s stop_reason=%s blocks=%d kinds=%s",
			msg.ID,
			msg.Role,
			msg.StopReason,
			len(msg.Content),
			strings.Join(blockKinds, ","),
		))
	}
	return out
}

// nextTurnReady is an internal message to create the next turn after persistence.
type nextTurnReady struct {
	roundID           domain.MessageID
	messageID         domain.MessageID
	results           []domaintools.Result
	messages          []domain.Message
	toolResultMessage domain.Message
}

// handleNextTurnReady creates and starts the next turn.
func (m *Model) handleNextTurnReady(msg nextTurnReady) tea.Cmd {
	// Create input with tool results (empty text)
	input := msgs.UserSubmittedInput{
		ToolResults: msg.results,
	}

	nextTurn := turn.New(
		m.theme,
		m.conversationID,
		m.accountID,
		msg.messageID,
		input,
		m.width,
		m.db,
		m.chatClient,
		m.toolRegistry,
		m.scope,
	)

	m.turns = append(m.turns, nextTurn)
	startStream := nextTurn.StartStream(msg.messages, nil)
	notifyPersist := func() tea.Msg {
		return msgs.ToolResultMessagePersisted{Message: msg.toolResultMessage}
	}
	return tea.Batch(startStream, notifyPersist)
}
