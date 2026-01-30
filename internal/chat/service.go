// Package chat provides chat domain services.
package chat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
)

// Service handles chat operations.
type Service struct {
	db     sqlite.Database
	logger log.Logger
}

// NewService creates a new chat service.
func NewService(db sqlite.Database, logger log.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// Message represents a chat message.
type Message struct {
	ID             string
	AccountID      string
	ConversationID string
	Content        string
	Role           string
	CreatedAt      time.Time
}

// Conversation represents a chat conversation.
type Conversation struct {
	ID        string
	AccountID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SendMessage sends a user message in a conversation.
// If conversationID is empty, a new conversation is created.
// Returns the conversation ID (useful when a new one was created).
func (s *Service) SendMessage(ctx context.Context, accountID, conversationID, text string) (string, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Ensure we have a conversation
	if conversationID == "" {
		var err error
		conversationID, err = s.createConversation(ctx, accountID, now)
		if err != nil {
			return "", err
		}
	}

	// Insert the message
	msgID := uuid.New().String()
	role := "user"
	err := s.db.Queries().InsertMessage(ctx, sqlite.InsertMessageParams{
		ID:             &msgID,
		AccountID:      &accountID,
		Content:        &text,
		ConversationID: &conversationID,
		CreatedAt:      &now,
		Role:           &role,
	})
	if err != nil {
		s.logger.Error("failed to insert message", "error", err)
		return "", err
	}

	s.logger.Debug("sent message", "messageID", msgID, "conversationID", conversationID)
	return conversationID, nil
}

// createConversation creates a new conversation.
func (s *Service) createConversation(ctx context.Context, accountID, now string) (string, error) {
	convID := uuid.New().String()
	err := s.db.Queries().InsertConversation(ctx, sqlite.InsertConversationParams{
		ID:        &convID,
		AccountID: &accountID,
		CreatedAt: &now,
		UpdatedAt: &now,
	})
	if err != nil {
		s.logger.Error("failed to create conversation", "error", err)
		return "", err
	}

	s.logger.Debug("created conversation", "conversationID", convID)
	return convID, nil
}

// GetConversation gets a conversation by ID.
func (s *Service) GetConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	conv, err := s.db.Queries().GetConversation(ctx, &conversationID)
	if err != nil {
		return nil, err
	}

	return s.conversationFromRow(conv), nil
}

// GetLatestConversation gets the most recent conversation for an account.
func (s *Service) GetLatestConversation(ctx context.Context, accountID string) (*Conversation, error) {
	conv, err := s.db.Queries().GetLatestConversationByAccount(ctx, &accountID)
	if err != nil {
		return nil, err
	}

	return s.conversationFromRow(conv), nil
}

func (s *Service) conversationFromRow(conv sqlite.Conversation) *Conversation {
	c := &Conversation{}
	if conv.ID != nil {
		c.ID = *conv.ID
	}
	if conv.AccountID != nil {
		c.AccountID = *conv.AccountID
	}
	if conv.CreatedAt != nil {
		c.CreatedAt, _ = time.Parse(time.RFC3339, *conv.CreatedAt)
	}
	if conv.UpdatedAt != nil {
		c.UpdatedAt, _ = time.Parse(time.RFC3339, *conv.UpdatedAt)
	}
	return c
}
