package chattest

import (
	"context"

	"github.com/usetero/cli/internal/api/chatclient"
	"github.com/usetero/cli/internal/domain"
)

// MockClient is a mock implementation of chat.Client for testing.
type MockClient struct {
	StreamFunc          func(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) (*chat.StreamResult, error)
	StreamSnapshotsFunc func(ctx context.Context, req chat.Request, onSnapshot func(chat.StreamSnapshot)) (*chat.StreamResult, error)
	SetAccountFunc      func(accountID domain.AccountID)
	WithAccountFunc     func(accountID domain.AccountID) chat.Client
}

var _ chat.Client = (*MockClient)(nil)

func (m *MockClient) Stream(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) (*chat.StreamResult, error) {
	if m.StreamFunc != nil {
		return m.StreamFunc(ctx, req, onMessage)
	}
	if m.StreamSnapshotsFunc != nil {
		return m.StreamSnapshots(ctx, req, func(s chat.StreamSnapshot) {
			if onMessage != nil {
				onMessage(s.Message)
			}
		})
	}
	return &chat.StreamResult{}, nil
}

func (m *MockClient) StreamSnapshots(ctx context.Context, req chat.Request, onSnapshot func(chat.StreamSnapshot)) (*chat.StreamResult, error) {
	if m.StreamSnapshotsFunc != nil {
		return m.StreamSnapshotsFunc(ctx, req, onSnapshot)
	}
	if m.StreamFunc != nil {
		return m.StreamFunc(ctx, req, func(msg *domain.Message) {
			if onSnapshot != nil {
				onSnapshot(chat.StreamSnapshot{Message: msg, Status: chat.StreamStatusStreaming})
			}
		})
	}
	return &chat.StreamResult{}, nil
}

func (m *MockClient) SetAccountID(accountID domain.AccountID) {
	if m.SetAccountFunc != nil {
		m.SetAccountFunc(accountID)
	}
}

func (m *MockClient) WithAccountID(accountID domain.AccountID) chat.Client {
	if m.WithAccountFunc != nil {
		return m.WithAccountFunc(accountID)
	}
	if m.SetAccountFunc != nil {
		m.SetAccountFunc(accountID)
	}
	return m
}
