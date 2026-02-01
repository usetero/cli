package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockConversations implements api.Conversations for testing.
type MockConversations struct {
	CreateFunc func(ctx context.Context, workspaceID, title string) (*api.Conversation, error)
	UpdateFunc func(ctx context.Context, id, title string) (*api.Conversation, error)
	DeleteFunc func(ctx context.Context, id string) error
}

func (m *MockConversations) Create(ctx context.Context, workspaceID, title string) (*api.Conversation, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, workspaceID, title)
	}
	return nil, nil
}

func (m *MockConversations) Update(ctx context.Context, id, title string) (*api.Conversation, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, title)
	}
	return nil, nil
}

func (m *MockConversations) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
