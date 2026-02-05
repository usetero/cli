package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Conversations provides access to conversations.
type Conversations interface {
	Create(ctx context.Context, id uuid.UUID, workspaceID domain.WorkspaceID, title string) (*domain.Conversation, error)
	Update(ctx context.Context, id domain.ConversationID, title string) (*domain.Conversation, error)
	Delete(ctx context.Context, id domain.ConversationID) error
}

// ConversationService handles conversation-related API operations.
type ConversationService struct {
	client Client
	scope  log.Scope
}

// Ensure ConversationService implements Conversations.
var _ Conversations = (*ConversationService)(nil)

// NewConversationService creates a new conversation service.
func NewConversationService(client Client, scope log.Scope) *ConversationService {
	return &ConversationService{
		client: client,
		scope:  scope.Child("conversations"),
	}
}

// Create creates a new conversation with the given client-provided ID.
func (s *ConversationService) Create(ctx context.Context, id uuid.UUID, workspaceID domain.WorkspaceID, title string) (*domain.Conversation, error) {
	s.scope.Debug("creating conversation via API", "id", id.String(), "workspaceID", workspaceID.String(), "title", title)

	input := gen.CreateConversationInput{
		Id:          ptr(id.String()),
		WorkspaceID: workspaceID.String(),
		Title:       ptr(title),
	}

	resp, err := s.client.CreateConversation(ctx, input)
	if err != nil {
		s.scope.Error("failed to create conversation", "error", err)
		if classified := classifyError(err); classified != nil {
			return nil, fmt.Errorf("create conversation %s: %w", id, classified)
		}
		return nil, err
	}

	conversation := &domain.Conversation{
		ID:          domain.ConversationID(resp.CreateConversation.Id),
		WorkspaceID: workspaceID,
		Title:       title,
	}

	s.scope.Debug("created conversation via API", "id", conversation.ID)
	return conversation, nil
}

// Update updates a conversation's title.
func (s *ConversationService) Update(ctx context.Context, id domain.ConversationID, title string) (*domain.Conversation, error) {
	s.scope.Debug("updating conversation via API", "id", id.String(), "title", title)

	input := gen.UpdateConversationInput{
		Title: ptr(title),
	}

	resp, err := s.client.UpdateConversation(ctx, id.String(), input)
	if err != nil {
		s.scope.Error("failed to update conversation", "error", err, "id", id)
		if classified := classifyError(err); classified != nil {
			return nil, fmt.Errorf("update conversation %s: %w", id, classified)
		}
		return nil, err
	}

	conversation := &domain.Conversation{
		ID:    domain.ConversationID(resp.UpdateConversation.Id),
		Title: deref(resp.UpdateConversation.Title),
	}

	s.scope.Debug("updated conversation via API", "id", conversation.ID)
	return conversation, nil
}

// Delete deletes a conversation.
// Returns ErrNotFound (via errors.Is) if the conversation does not exist.
func (s *ConversationService) Delete(ctx context.Context, id domain.ConversationID) error {
	s.scope.Debug("deleting conversation via API", "id", id.String())

	_, err := s.client.DeleteConversation(ctx, id.String())
	if err != nil {
		s.scope.Error("failed to delete conversation", "error", err, "id", id)
		if classified := classifyError(err); classified != nil {
			return errors.Join(fmt.Errorf("delete conversation %s", id), classified)
		}
		return err
	}

	s.scope.Debug("deleted conversation via API", "id", id.String())
	return nil
}
