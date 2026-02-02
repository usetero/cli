package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/pkg/client"
)

// Conversations provides access to conversations.
type Conversations interface {
	Create(ctx context.Context, id uuid.UUID, workspaceID, title string) (*Conversation, error)
	Update(ctx context.Context, id, title string) (*Conversation, error)
	Delete(ctx context.Context, id string) error
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

// Create creates a new conversation with the given client-provided ID.
func (s *ConversationService) Create(ctx context.Context, id uuid.UUID, workspaceID, title string) (*Conversation, error) {
	s.logger.Debug("creating conversation via API", "id", id.String(), "workspaceID", workspaceID, "title", title)

	input := client.CreateConversationInput{
		Id:          id.String(),
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

// Update updates a conversation's title.
func (s *ConversationService) Update(ctx context.Context, id, title string) (*Conversation, error) {
	s.logger.Debug("updating conversation via API", "id", id, "title", title)

	input := client.UpdateConversationInput{
		Title: title,
	}

	resp, err := s.client.UpdateConversation(ctx, id, input)
	if err != nil {
		s.logger.Error("failed to update conversation", "error", err, "id", id)
		return nil, err
	}

	conversation := &Conversation{
		ID:    resp.UpdateConversation.Id,
		Title: resp.UpdateConversation.Title,
	}

	s.logger.Debug("updated conversation via API", "id", conversation.ID)
	return conversation, nil
}

// Delete deletes a conversation.
func (s *ConversationService) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting conversation via API", "id", id)

	_, err := s.client.DeleteConversation(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete conversation", "error", err, "id", id)
		return err
	}

	s.logger.Debug("deleted conversation via API", "id", id)
	return nil
}
