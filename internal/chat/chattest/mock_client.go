package chattest

import (
	"context"

	"github.com/usetero/cli/internal/chat"
)

// MockClient is a mock implementation of chat.Client for testing.
type MockClient struct {
	SendFunc       func(ctx context.Context, req chat.Request, handler chat.Handler) error
	SetAccountFunc func(accountID string)
}

var _ chat.Client = (*MockClient)(nil)

func (m *MockClient) Send(ctx context.Context, req chat.Request, handler chat.Handler) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, req, handler)
	}
	return nil
}

func (m *MockClient) SetAccountID(accountID string) {
	if m.SetAccountFunc != nil {
		m.SetAccountFunc(accountID)
	}
}
