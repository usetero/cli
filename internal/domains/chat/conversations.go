package chat

import (
	"context"
	"strings"
	"time"
)

type ConversationID string

type Conversation struct {
	ID        ConversationID
	Title     *string
	CreatedAt time.Time
}

// ConversationCreate is the conversation creation mutation input.
type ConversationCreate struct {
	Title *string
}

// Validate normalizes conversation create input.
func (c ConversationCreate) Validate() (ConversationCreate, error) {
	if c.Title == nil {
		return c, nil
	}
	title := strings.TrimSpace(*c.Title)
	if title == "" {
		c.Title = nil
		return c, nil
	}
	c.Title = &title
	return c, nil
}

// ConversationService is the domain contract for conversation operations.
type ConversationService interface {
	Create(ctx context.Context, create ConversationCreate) (ConversationID, error)
	Delete(ctx context.Context, id ConversationID) error
	List(ctx context.Context) ([]Conversation, error)
}
