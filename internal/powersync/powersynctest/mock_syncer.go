package powersynctest

import (
	"context"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// MockSyncer is a test double for powersync.Syncer.
type MockSyncer struct {
	StartFunc   func(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error
	StopFunc    func()
	StateFunc   func() powersync.State
	IsReadyFunc func() bool
}

var _ powersync.Syncer = (*MockSyncer)(nil)

// NewMockSyncer creates a MockSyncer with sensible defaults.
func NewMockSyncer() *MockSyncer {
	return &MockSyncer{}
}

func (m *MockSyncer) Start(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, db, accountID, onFirstSync)
	}
	return nil
}

func (m *MockSyncer) Stop() {
	if m.StopFunc != nil {
		m.StopFunc()
	}
}

func (m *MockSyncer) State() powersync.State {
	if m.StateFunc != nil {
		return m.StateFunc()
	}
	return powersync.NewDisconnected()
}

func (m *MockSyncer) IsReady() bool {
	if m.IsReadyFunc != nil {
		return m.IsReadyFunc()
	}
	return false
}
