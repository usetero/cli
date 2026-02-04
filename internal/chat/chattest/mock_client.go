package chattest

import (
	"context"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
)

// MockClient is a mock implementation of chat.Client for testing.
type MockClient struct {
	StreamFunc     func(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) error
	SetAccountFunc func(accountID domain.AccountID)
}

var _ chat.Client = (*MockClient)(nil)

func (m *MockClient) Stream(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) error {
	if m.StreamFunc != nil {
		return m.StreamFunc(ctx, req, onMessage)
	}
	return nil
}

func (m *MockClient) SetAccountID(accountID domain.AccountID) {
	if m.SetAccountFunc != nil {
		m.SetAccountFunc(accountID)
	}
}
