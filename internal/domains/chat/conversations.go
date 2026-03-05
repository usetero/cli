package chat

import (
	"context"
	"time"
)

type ConversationID string

type Conversation struct {
	ID        ConversationID
	Title     *string
	CreatedAt time.Time
}

// ConversationService is the domain contract for conversation operations.
type ConversationService interface {
	Create(ctx context.Context, title *string) (ConversationID, error)
	Delete(ctx context.Context, id ConversationID) error
	List(ctx context.Context) ([]Conversation, error)
}
