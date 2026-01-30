// Package powersynctest provides test doubles for the powersync package.
package powersynctest

import (
	"context"

	"github.com/usetero/cli/internal/powersync"
)

// MockStreamer is a test double for powersync.Streamer.
type MockStreamer struct {
	// ConnectFunc is called when Connect is invoked.
	ConnectFunc func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error

	// Token is the current token set via SetToken.
	Token string

	// ConnectCalls records the number of times Connect was called.
	ConnectCalls int
}

// Ensure MockStreamer implements powersync.Streamer.
var _ powersync.Streamer = (*MockStreamer)(nil)

// Connect implements powersync.Streamer.
func (m *MockStreamer) Connect(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
	m.ConnectCalls++
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx, req, handler)
	}
	return nil
}

// SetToken implements powersync.Streamer.
func (m *MockStreamer) SetToken(token string) {
	m.Token = token
}

// NewMockStreamerFactory returns a stream factory that always returns the given mock.
// Useful for injecting a pre-configured mock into Sync.
func NewMockStreamerFactory(mock *MockStreamer) func(endpoint, token string) powersync.Streamer {
	return func(endpoint, token string) powersync.Streamer {
		mock.Token = token
		return mock
	}
}
