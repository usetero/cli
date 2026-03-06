package chat

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/domains/validation"
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

// UserMessageCreate is the user message creation mutation input.
type UserMessageCreate struct {
	ConversationID ConversationID `label:"conversation id" validate:"required"`
	Content        string         `label:"content" validate:"required,notblank"`
}

// Validate validates user message create input.
func (c UserMessageCreate) Validate() (UserMessageCreate, error) {
	if err := validation.Struct(c); err != nil {
		return UserMessageCreate{}, err
	}
	return c, nil
}

// MessageService is the domain contract for message operations.
type MessageService interface {
	CreateUserMessage(ctx context.Context, create UserMessageCreate) (MessageID, error)
	Delete(ctx context.Context, messageID MessageID) error
	ListByConversation(ctx context.Context, conversationID ConversationID) ([]Message, error)
}
