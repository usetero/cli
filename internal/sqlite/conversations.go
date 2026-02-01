package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Conversations provides type-safe access to conversations.
type Conversations interface {
	Create(ctx context.Context, accountID string) (string, error)
	List(ctx context.Context, accountID string) ([]gen.Conversation, error)
	Get(ctx context.Context, id string) (gen.Conversation, error)
}

// conversationsImpl implements Conversations.
type conversationsImpl struct {
	queries *gen.Queries
}

// Create creates a new conversation and returns its ID.
func (c *conversationsImpl) Create(ctx context.Context, accountID string) (string, error) {
	convID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	err := c.queries.InsertConversation(ctx, gen.InsertConversationParams{
		ID:        &convID,
		AccountID: &accountID,
		CreatedAt: &now,
		UpdatedAt: &now,
	})
	if err != nil {
		return "", err
	}

	return convID, nil
}

// List returns all conversations for an account.
func (c *conversationsImpl) List(ctx context.Context, accountID string) ([]gen.Conversation, error) {
	return c.queries.ListConversationsByAccount(ctx, &accountID)
}

// Get returns a conversation by ID.
func (c *conversationsImpl) Get(ctx context.Context, id string) (gen.Conversation, error) {
	return c.queries.GetConversation(ctx, &id)
}
