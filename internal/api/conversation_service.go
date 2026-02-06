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

// CreateConversationInput contains the fields for creating a conversation.
type CreateConversationInput struct {
	ID          uuid.UUID
	WorkspaceID domain.WorkspaceID
	Title       string
}

// UpdateConversationInput contains the fields that can be updated on a conversation.
// Fields are pointers — nil means "don't change", non-nil means "set to this value".
type UpdateConversationInput struct {
	Title *string
}

// Conversations provides access to conversations.
type Conversations interface {
	Create(ctx context.Context, input CreateConversationInput) (*domain.Conversation, error)
	Update(ctx context.Context, id domain.ConversationID, input UpdateConversationInput) (*domain.Conversation, error)
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
func (s *ConversationService) Create(ctx context.Context, input CreateConversationInput) (*domain.Conversation, error) {
	s.scope.Debug("creating conversation via API", "id", input.ID.String(), "workspaceID", input.WorkspaceID.String(), "title", input.Title)

	genInput := gen.CreateConversationInput{
		Id:          ptr(input.ID.String()),
		WorkspaceID: input.WorkspaceID.String(),
		Title:       ptr(input.Title),
	}

	resp, err := s.client.CreateConversation(ctx, genInput)
	if err != nil {
		s.scope.Error("failed to create conversation", "error", err)
		if classified := classifyError(err); classified != nil {
			return nil, fmt.Errorf("create conversation %s: %w", input.ID, classified)
		}
		return nil, err
	}

	conversation := &domain.Conversation{
		ID:          domain.ConversationID(resp.CreateConversation.Id),
		WorkspaceID: input.WorkspaceID,
		Title:       input.Title,
	}

	s.scope.Debug("created conversation via API", "id", conversation.ID)
	return conversation, nil
}

// Update updates a conversation.
func (s *ConversationService) Update(ctx context.Context, id domain.ConversationID, input UpdateConversationInput) (*domain.Conversation, error) {
	s.scope.Debug("updating conversation via API", "id", id.String())

	genInput := gen.UpdateConversationInput{}
	if input.Title != nil {
		if *input.Title == "" {
			genInput.ClearTitle = ptr(true)
		} else {
			genInput.Title = input.Title
		}
	}

	resp, err := s.client.UpdateConversation(ctx, id.String(), genInput)
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
