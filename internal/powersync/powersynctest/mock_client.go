// Package powersynctest provides test doubles for the powersync package.
package powersynctest

import (
	"context"
	"io"
	"log/slog"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
)

// MockClient is a test double for powersync.Client.
type MockClient struct {
	// SyncStreamFunc is called when SyncStream is invoked.
	SyncStreamFunc func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error

	// GetWriteCheckpointFunc is called when GetWriteCheckpoint is invoked.
	GetWriteCheckpointFunc func(ctx context.Context, clientID string) (string, error)

	// Token is the current token set via SetToken.
	Token string

	// SyncStreamCalls records the number of times SyncStream was called.
	SyncStreamCalls int

	// GetWriteCheckpointCalls records the number of times GetWriteCheckpoint was called.
	GetWriteCheckpointCalls int
}

// Ensure MockClient implements powersync.Client.
var _ powersync.Client = (*MockClient)(nil)

// NewMockClient creates a new MockClient with sensible defaults.
func NewMockClient() *MockClient {
	return &MockClient{
		SyncStreamFunc: func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
}

// SyncStream implements powersync.Client.
func (m *MockClient) SyncStream(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
	m.SyncStreamCalls++
	if m.SyncStreamFunc != nil {
		return m.SyncStreamFunc(ctx, req, handler)
	}
	return nil
}

// GetWriteCheckpoint implements powersync.Client.
func (m *MockClient) GetWriteCheckpoint(ctx context.Context, clientID string) (string, error) {
	m.GetWriteCheckpointCalls++
	if m.GetWriteCheckpointFunc != nil {
		return m.GetWriteCheckpointFunc(ctx, clientID)
	}
	return "1", nil
}

// SetToken implements powersync.Client.
func (m *MockClient) SetToken(token string) {
	m.Token = token
}

// NewMockClientFactory returns a client factory that always returns the given mock.
func NewMockClientFactory(mock *MockClient) func(endpoint string) powersync.Client {
	return func(endpoint string) powersync.Client {
		return mock
	}
}

// NewSyncWithMockClient creates a Sync with a mock client for testing.
func NewSyncWithMockClient(endpoint string, tokenRefresher powersync.TokenRefresher, mock *MockClient) *powersync.Sync {
	sync := powersync.NewSync(endpoint, tokenRefresher, discardLogger())
	sync.SetClientFactory(NewMockClientFactory(mock))
	return sync
}

func discardLogger() log.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
