package chat

import (
	tea "charm.land/bubbletea/v2"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
)

// handleUserInput creates conversation if needed, then persists the user message.
func (m *Model) handleUserInput(input msgs.UserSubmittedInput) tea.Cmd {
	if len(input.Text) > 0 {
		m.scope.Info("user submitted text", "text_length", len(input.Text))
	} else {
		m.scope.Info("user submitted tool results", "count", len(input.ToolResults))
	}

	// Cancel any in-flight round before starting a new one.
	_, cancelCmd := m.CancelActiveRound()

	// If no conversation yet, create one first (only for text input).
	if m.conversationID == "" {
		return tea.Batch(cancelCmd, m.createConversation(input))
	}

	return tea.Batch(cancelCmd, m.persistUserMessage(input))
}

// createConversation starts a new ephemeral conversation. Chat is not
// persisted, so the conversation ID is minted locally for this session only.
func (m *Model) createConversation(input msgs.UserSubmittedInput) tea.Cmd {
	return func() tea.Msg {
		return conversationCreated{
			conversationID: domain.NewConversationID(),
			input:          input,
		}
	}
}

// conversationCreated is fired after conversation is created.
type conversationCreated struct {
	conversationID domain.ConversationID
	input          msgs.UserSubmittedInput
}

// persistUserMessage appends the user message to the in-memory session. Chat
// is ephemeral, so the message ID is minted locally and nothing is stored.
func (m *Model) persistUserMessage(input msgs.UserSubmittedInput) tea.Cmd {
	return func() tea.Msg {
		var domainResults []domain.ToolResult
		if len(input.ToolResults) > 0 {
			// Convert typed results to domain format at the boundary.
			domainResults = make([]domain.ToolResult, len(input.ToolResults))
			for i, r := range input.ToolResults {
				domainResults[i] = domain.ToolResult{
					ToolUseID: r.ToolUseID,
					IsError:   r.IsError(),
					Content:   r.ToMap(),
				}
				if r.Error != nil {
					domainResults[i].Error = r.Error.Message
				}
			}
		}

		msgID := domain.NewMessageID()
		if m.session == nil {
			m.session = corechat.NewSession(m.conversationID, nil)
		}
		if len(domainResults) > 0 {
			m.session.AppendUserToolResultsMessage(msgID, domainResults)
		} else {
			m.session.AppendUserTextMessage(msgID, input.Text)
		}
		messages := m.session.Messages()

		return userMessagePersisted{
			conversationID: m.conversationID,
			messageID:      msgID,
			input:          input,
			messages:       messages,
		}
	}
}

// userMessagePersisted is fired after the user message is appended to the session.
type userMessagePersisted struct {
	conversationID domain.ConversationID
	messageID      domain.MessageID
	input          msgs.UserSubmittedInput
	messages       []domain.Message
}

// handlePersistedMessage starts the turn after the user message is persisted.
func (m *Model) handlePersistedMessage(msg userMessagePersisted) tea.Cmd {
	m.scope.Info("turn started", "conversation_id", msg.conversationID, "user_message_id", msg.messageID)

	return m.messageList.StartTurn(
		msg.conversationID,
		m.account.ID,
		msg.messageID,
		msg.input,
		msg.messages,
		nil,
	)
}
