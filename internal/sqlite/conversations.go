package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Conversations provides type-safe access to conversations.
type Conversations interface {
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, accountID domain.AccountID) (domain.ConversationID, error)
	UpdateTitle(ctx context.Context, id domain.ConversationID, title string) error
	List(ctx context.Context, accountID domain.AccountID) ([]gen.Conversation, error)
	Get(ctx context.Context, id domain.ConversationID) (gen.Conversation, error)
}

// conversationsImpl implements Conversations.
type conversationsImpl struct {
	read  *gen.Queries // read pool — Count, List, Get
	write *gen.Queries // write pool — Create, UpdateTitle
}

// Count returns the total number of conversations.
func (c *conversationsImpl) Count(ctx context.Context) (int64, error) {
	count, err := c.read.CountConversations(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count conversations")
	}
	return count, nil
}

// Create creates a new conversation and returns its ID.
func (c *conversationsImpl) Create(ctx context.Context, accountID domain.AccountID) (domain.ConversationID, error) {
	convID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	accountIDStr := accountID.String()

	err := c.write.InsertConversation(ctx, gen.InsertConversationParams{
		ID:        &convID,
		AccountID: &accountIDStr,
		CreatedAt: &now,
	})
	if err != nil {
		return "", WrapSQLiteError(err, "insert conversation")
	}

	return domain.ConversationID(convID), nil
}

// UpdateTitle sets the title on a conversation.
func (c *conversationsImpl) UpdateTitle(ctx context.Context, id domain.ConversationID, title string) error {
	idStr := id.String()
	err := c.write.UpdateConversationTitle(ctx, gen.UpdateConversationTitleParams{
		Title: &title,
		ID:    &idStr,
	})
	return WrapSQLiteError(err, "update conversation title")
}

// List returns all conversations for an account.
func (c *conversationsImpl) List(ctx context.Context, accountID domain.AccountID) ([]gen.Conversation, error) {
	accountIDStr := accountID.String()
	return c.read.ListConversationsByAccount(ctx, &accountIDStr)
}

// Get returns a conversation by ID.
func (c *conversationsImpl) Get(ctx context.Context, id domain.ConversationID) (gen.Conversation, error) {
	idStr := id.String()
	return c.read.GetConversation(ctx, &idStr)
}
