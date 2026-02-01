package api

import (
	"context"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/pkg/client"
)

// Conversations provides access to conversations.
type Conversations interface {
	Create(ctx context.Context, workspaceID, title string) (*Conversation, error)
}

// ConversationService handles conversation-related API operations.
type ConversationService struct {
	client Client
	logger log.Logger
}

// Ensure ConversationService implements Conversations.
var _ Conversations = (*ConversationService)(nil)

// NewConversationService creates a new conversation service.
func NewConversationService(client Client, logger log.Logger) *ConversationService {
	return &ConversationService{
		client: client,
		logger: logger,
	}
}

// Conversation is the domain model for a conversation.
type Conversation struct {
	ID          string
	WorkspaceID string
	Title       string
}

// Create creates a new conversation.
func (s *ConversationService) Create(ctx context.Context, workspaceID, title string) (*Conversation, error) {
	s.logger.Debug("creating conversation via API", "workspaceID", workspaceID, "title", title)

	input := client.CreateConversationInput{
		WorkspaceID: workspaceID,
		Title:       title,
	}

	resp, err := s.client.CreateConversation(ctx, input)
	if err != nil {
		s.logger.Error("failed to create conversation", "error", err)
		return nil, err
	}

	conversation := &Conversation{
		ID:          resp.CreateConversation.Id,
		WorkspaceID: workspaceID,
		Title:       title,
	}

	s.logger.Debug("created conversation via API", "id", conversation.ID)
	return conversation, nil
}
