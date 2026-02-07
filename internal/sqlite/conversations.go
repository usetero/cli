package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Conversations provides type-safe access to conversations.
type Conversations interface {
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, accountID, workspaceID string) (string, error)
	UpdateTitle(ctx context.Context, id, title string) error
	List(ctx context.Context, accountID string) ([]gen.Conversation, error)
	Get(ctx context.Context, id string) (gen.Conversation, error)
}

// conversationsImpl implements Conversations.
type conversationsImpl struct {
	queries *gen.Queries
}

// Count returns the total number of conversations.
func (c *conversationsImpl) Count(ctx context.Context) (int64, error) {
	count, err := c.queries.CountConversations(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count conversations")
	}
	return count, nil
}

// Create creates a new conversation and returns its ID.
func (c *conversationsImpl) Create(ctx context.Context, accountID, workspaceID string) (string, error) {
	convID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	err := c.queries.InsertConversation(ctx, gen.InsertConversationParams{
		ID:          &convID,
		AccountID:   &accountID,
		WorkspaceID: &workspaceID,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	})
	if err != nil {
		return "", WrapSQLiteError(err, "insert conversation")
	}

	return convID, nil
}

// UpdateTitle sets the title on a conversation.
func (c *conversationsImpl) UpdateTitle(ctx context.Context, id, title string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	err := c.queries.UpdateConversationTitle(ctx, gen.UpdateConversationTitleParams{
		Title:     &title,
		UpdatedAt: &now,
		ID:        &id,
	})
	return WrapSQLiteError(err, "update conversation title")
}

// List returns all conversations for an account.
func (c *conversationsImpl) List(ctx context.Context, accountID string) ([]gen.Conversation, error) {
	return c.queries.ListConversationsByAccount(ctx, &accountID)
}

// Get returns a conversation by ID.
func (c *conversationsImpl) Get(ctx context.Context, id string) (gen.Conversation, error) {
	return c.queries.GetConversation(ctx, &id)
}
