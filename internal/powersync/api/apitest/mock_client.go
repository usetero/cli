// Package apitest provides test doubles for the powersync/api package.
package apitest

import (
	"context"

	"github.com/usetero/cli/internal/powersync/api"
)

// MockClient is a test double for api.Client.
type MockClient struct {
	// SyncStreamFunc is called when SyncStream is invoked.
	SyncStreamFunc func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error

	// GetWriteCheckpointFunc is called when GetWriteCheckpoint is invoked.
	GetWriteCheckpointFunc func(ctx context.Context, clientID string) (string, error)

	// Token is the current token set via SetToken.
	Token string

	// SyncStreamCalls records the number of times SyncStream was called.
	SyncStreamCalls int

	// GetWriteCheckpointCalls records the number of times GetWriteCheckpoint was called.
	GetWriteCheckpointCalls int
}

// Ensure MockClient implements api.Client.
var _ api.Client = (*MockClient)(nil)

// NewMockClient creates a new MockClient with sensible defaults.
func NewMockClient() *MockClient {
	return &MockClient{
		SyncStreamFunc: func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
}

// SyncStream implements api.Client.
func (m *MockClient) SyncStream(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
	m.SyncStreamCalls++
	if m.SyncStreamFunc != nil {
		return m.SyncStreamFunc(ctx, req, handler)
	}
	return nil
}

// GetWriteCheckpoint implements api.Client.
func (m *MockClient) GetWriteCheckpoint(ctx context.Context, clientID string) (string, error) {
	m.GetWriteCheckpointCalls++
	if m.GetWriteCheckpointFunc != nil {
		return m.GetWriteCheckpointFunc(ctx, clientID)
	}
	return "1", nil
}

// SetToken implements api.Client.
func (m *MockClient) SetToken(token string) {
	m.Token = token
}

// NewMockClientFactory returns a client factory that always returns the given mock.
func NewMockClientFactory(mock *MockClient) func(endpoint string) api.Client {
	return func(endpoint string) api.Client {
		return mock
	}
}
