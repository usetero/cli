package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockConversations implements api.Conversations for testing.
type MockConversations struct {
	CreateFunc func(ctx context.Context, workspaceID, title string) (*api.Conversation, error)
}

func (m *MockConversations) Create(ctx context.Context, workspaceID, title string) (*api.Conversation, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, workspaceID, title)
	}
	return nil, nil
}
