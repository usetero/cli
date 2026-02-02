package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Role identifies who sent a message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Messages provides type-safe access to messages.
type Messages interface {
	CreateUserMessage(ctx context.Context, accountID, conversationID, text string) (string, error)
	CreateAssistantMessage(ctx context.Context, accountID, conversationID, model string) (string, error)
	UpdateContent(ctx context.Context, id, content string) error
	UpdateMeta(ctx context.Context, id, model, stopReason string) error
	List(ctx context.Context, conversationID string) ([]gen.Message, error)
}

// messagesImpl implements Messages.
type messagesImpl struct {
	queries *gen.Queries
}

// CreateUserMessage creates a user message with properly encoded content.
// Returns the new message ID.
func (m *messagesImpl) CreateUserMessage(ctx context.Context, accountID, conversationID, text string) (string, error) {
	msgID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	role := string(RoleUser)

	content, err := block.EncodeText(text)
	if err != nil {
		return "", err
	}

	err = m.queries.InsertMessage(ctx, gen.InsertMessageParams{
		ID:             &msgID,
		AccountID:      &accountID,
		ConversationID: &conversationID,
		Content:        &content,
		CreatedAt:      &now,
		Role:           &role,
	})
	if err != nil {
		return "", WrapSQLiteError(err, "insert user message")
	}

	return msgID, nil
}

// CreateAssistantMessage creates an empty assistant message placeholder.
// Returns the new message ID. Content is added via UpdateContent as it streams in.
func (m *messagesImpl) CreateAssistantMessage(ctx context.Context, accountID, conversationID, model string) (string, error) {
	msgID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	role := string(RoleAssistant)
	content := "[]" // Empty JSON array

	err := m.queries.InsertMessage(ctx, gen.InsertMessageParams{
		ID:             &msgID,
		AccountID:      &accountID,
		ConversationID: &conversationID,
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
func (m *messagesImpl) UpdateContent(ctx context.Context, id, content string) error {
	err := m.queries.UpdateMessageContent(ctx, gen.UpdateMessageContentParams{
		ID:      &id,
		Content: &content,
	})
	return WrapSQLiteError(err, "update message content")
}

// UpdateMeta updates the model and stop_reason of a message.
func (m *messagesImpl) UpdateMeta(ctx context.Context, id, model, stopReason string) error {
	err := m.queries.UpdateMessageMeta(ctx, gen.UpdateMessageMetaParams{
		ID:         &id,
		Model:      &model,
		StopReason: &stopReason,
	})
	return WrapSQLiteError(err, "update message meta")
}

// List returns all messages for a conversation, ordered by creation time.
func (m *messagesImpl) List(ctx context.Context, conversationID string) ([]gen.Message, error) {
	return m.queries.ListMessagesByConversation(ctx, &conversationID)
}
