package apitest

import (
	"context"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/domain"
)

// MockConversations implements api.Conversations for testing.
type MockConversations struct {
	CreateFunc func(ctx context.Context, id uuid.UUID, workspaceID domain.WorkspaceID, title string) (*domain.Conversation, error)
	UpdateFunc func(ctx context.Context, id domain.ConversationID, title string) (*domain.Conversation, error)
	DeleteFunc func(ctx context.Context, id domain.ConversationID) error
}

// NewMockConversations creates a MockConversations with sensible defaults.
func NewMockConversations() *MockConversations {
	return &MockConversations{}
}

func (m *MockConversations) Create(ctx context.Context, id uuid.UUID, workspaceID domain.WorkspaceID, title string) (*domain.Conversation, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, id, workspaceID, title)
	}
	return nil, nil
}

func (m *MockConversations) Update(ctx context.Context, id domain.ConversationID, title string) (*domain.Conversation, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, title)
	}
	return nil, nil
}

func (m *MockConversations) Delete(ctx context.Context, id domain.ConversationID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
