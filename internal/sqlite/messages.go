package sqlite

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Messages provides type-safe access to messages.
type Messages interface {
	Get(ctx context.Context, id domain.MessageID) (*domain.Message, error)
	CreateUserMessage(ctx context.Context, accountID domain.AccountID, conversationID domain.ConversationID, text string) (domain.MessageID, error)
	CreateToolResultMessage(ctx context.Context, accountID domain.AccountID, conversationID domain.ConversationID, results []domain.ToolResult) (domain.MessageID, error)
	CreateAssistantMessage(ctx context.Context, accountID domain.AccountID, conversationID domain.ConversationID, model string) (domain.MessageID, error)
	UpdateContent(ctx context.Context, id domain.MessageID, content string) error
	UpdateMeta(ctx context.Context, id domain.MessageID, model, stopReason string) error
	List(ctx context.Context, conversationID domain.ConversationID) ([]domain.Message, error)
}

// messagesImpl implements Messages.
type messagesImpl struct {
	queries *gen.Queries
}

// Get retrieves a message by ID.
func (m *messagesImpl) Get(ctx context.Context, id domain.MessageID) (*domain.Message, error) {
	idStr := id.String()
	row, err := m.queries.GetMessage(ctx, &idStr)
	if err != nil {
		return nil, WrapSQLiteError(err, "get message")
	}
	return toMessage(row), nil
}

// CreateUserMessage creates a user message with properly encoded content.
// Returns the new message ID.
func (m *messagesImpl) CreateUserMessage(ctx context.Context, accountID domain.AccountID, conversationID domain.ConversationID, text string) (domain.MessageID, error) {
	msgID := domain.NewMessageID()
	msgIDStr := msgID.String()
	accountIDStr := accountID.String()
	convIDStr := conversationID.String()
	now := time.Now().UTC().Format(time.RFC3339)
	role := string(domain.RoleUser)

	content, err := domain.EncodeText(text)
	if err != nil {
		return "", err
	}

	err = m.queries.InsertMessage(ctx, gen.InsertMessageParams{
		ID:             &msgIDStr,
		AccountID:      &accountIDStr,
		ConversationID: &convIDStr,
		Content:        &content,
		CreatedAt:      &now,
		Role:           &role,
	})
	if err != nil {
		return "", WrapSQLiteError(err, "insert user message")
	}

	return msgID, nil
}

// CreateToolResultMessage creates a user message containing tool results.
func (m *messagesImpl) CreateToolResultMessage(ctx context.Context, accountID domain.AccountID, conversationID domain.ConversationID, results []domain.ToolResult) (domain.MessageID, error) {
	msgID := domain.NewMessageID()
	msgIDStr := msgID.String()
	accountIDStr := accountID.String()
	convIDStr := conversationID.String()
	now := time.Now().UTC().Format(time.RFC3339)
	role := string(domain.RoleUser)

	content, err := domain.EncodeToolResults(results)
	if err != nil {
		return "", err
	}

	err = m.queries.InsertMessage(ctx, gen.InsertMessageParams{
		ID:             &msgIDStr,
		AccountID:      &accountIDStr,
		ConversationID: &convIDStr,
		Content:        &content,
		CreatedAt:      &now,
		Role:           &role,
	})
	if err != nil {
		return "", WrapSQLiteError(err, "insert tool result message")
	}

	return msgID, nil
}

// CreateAssistantMessage creates an empty assistant message placeholder.
// Returns the new message ID. Content is added via UpdateContent as it streams in.
func (m *messagesImpl) CreateAssistantMessage(ctx context.Context, accountID domain.AccountID, conversationID domain.ConversationID, model string) (domain.MessageID, error) {
	msgID := domain.NewMessageID()
	msgIDStr := msgID.String()
	accountIDStr := accountID.String()
	convIDStr := conversationID.String()
	now := time.Now().UTC().Format(time.RFC3339)
	role := string(domain.RoleAssistant)
	content := "[]" // Empty JSON array

	err := m.queries.InsertMessage(ctx, gen.InsertMessageParams{
		ID:             &msgIDStr,
		AccountID:      &accountIDStr,
		ConversationID: &convIDStr,
		Content:        &content,
		CreatedAt:      &now,
		Model:          &model,
		Role:           &role,
	})
	if err != nil {
		return "", WrapSQLiteError(err, "insert assistant message")
	}

	return msgID, nil
}

// UpdateContent updates the content of a message.
func (m *messagesImpl) UpdateContent(ctx context.Context, id domain.MessageID, content string) error {
	idStr := id.String()
	err := m.queries.UpdateMessageContent(ctx, gen.UpdateMessageContentParams{
		ID:      &idStr,
		Content: &content,
	})
	return WrapSQLiteError(err, "update message content")
}

// UpdateMeta updates the model and stop_reason of a message.
func (m *messagesImpl) UpdateMeta(ctx context.Context, id domain.MessageID, model, stopReason string) error {
	idStr := id.String()
	err := m.queries.UpdateMessageMeta(ctx, gen.UpdateMessageMetaParams{
		ID:         &idStr,
		Model:      &model,
		StopReason: &stopReason,
	})
	return WrapSQLiteError(err, "update message meta")
}

// List returns all messages for a conversation, ordered by creation time.
func (m *messagesImpl) List(ctx context.Context, conversationID domain.ConversationID) ([]domain.Message, error) {
	convIDStr := conversationID.String()
	rows, err := m.queries.ListMessagesByConversation(ctx, &convIDStr)
	if err != nil {
		return nil, WrapSQLiteError(err, "list messages")
	}

	messages := make([]domain.Message, 0, len(rows))
	for _, row := range rows {
		if msg := toMessage(row); msg != nil {
			messages = append(messages, *msg)
		}
	}
	return messages, nil
}

// toMessage converts a gen.Message to a domain.Message.
// Returns nil if essential fields are missing.
func toMessage(row gen.Message) *domain.Message {
	if row.ID == nil || row.Role == nil {
		return nil
	}

	msg := &domain.Message{
		ID:   domain.MessageID(*row.ID),
		Role: domain.Role(*row.Role),
	}

	if row.ConversationID != nil {
		msg.ConversationID = domain.ConversationID(*row.ConversationID)
	}
	if row.Model != nil {
		msg.Model = *row.Model
	}
	if row.StopReason != nil {
		msg.StopReason = *row.StopReason
	}
	if row.CreatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *row.CreatedAt); err == nil {
			msg.CreatedAt = t
		}
	}
	if row.Content != nil {
		if blocks, err := domain.ParseBlocks(*row.Content); err == nil {
			msg.Content = blocks
		}
	}

	return msg
}
