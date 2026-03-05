package chat

import (
	"context"
	"time"
)

type MessageID string
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is the chat domain message model.
type Message struct {
	ID             MessageID
	ConversationID ConversationID
	Role           Role
	Content        string
	CreatedAt      time.Time
}

// MessageService is the domain contract for message operations.
type MessageService interface {
	CreateUserMessage(ctx context.Context, conversationID ConversationID, content string) (MessageID, error)
	Delete(ctx context.Context, messageID MessageID) error
	ListByConversation(ctx context.Context, conversationID ConversationID) ([]Message, error)
}
