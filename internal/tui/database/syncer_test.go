package database

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// mockPowerSync implements powersync.Syncer for testing.
type mockPowerSync struct {
	startFunc  func(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error
	status     powersync.Status
	syncStatus *powersync.SyncStatus
	lastError  error
	running    bool
	stopped    bool
}

func (m *mockPowerSync) Start(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error {
	m.running = true
	if m.startFunc != nil {
		return m.startFunc(ctx, db, accountID, onFirstSync)
	}
	return nil
}

func (m *mockPowerSync) Stop() {
	m.running = false
	m.stopped = true
}

func (m *mockPowerSync) Status() powersync.Status          { return m.status }
func (m *mockPowerSync) SyncStatus() *powersync.SyncStatus { return m.syncStatus }
func (m *mockPowerSync) LastError() error                  { return m.lastError }
func (m *mockPowerSync) IsRunning() bool                   { return m.running }

var _ powersync.Syncer = (*mockPowerSync)(nil)

func TestSyncer_Update(t *testing.T) {
	t.Parallel()

	t.Run("syncReadyMsg starts polling", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{status: powersync.StatusSyncing, running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))

		cmd := syncer.Update(syncReadyMsg{})

		if cmd == nil {
			t.Fatal("expected tick command")
		}
		if !syncer.waiting {
			t.Error("should be waiting after syncReadyMsg")
		}
	})

	t.Run("syncTickMsg sends SyncReadyMsg when firstSyncDone", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{status: powersync.StatusConnected, running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = true
		syncer.firstSyncDone = true

		cmd := syncer.Update(syncTickMsg{})

		if cmd == nil {
			t.Fatal("expected command")
		}

		msg := cmd()
		if _, ok := msg.(powersync.SyncReadyMsg); !ok {
			t.Errorf("expected SyncReadyMsg, got %T", msg)
		}
		if syncer.waiting {
			t.Error("should not be waiting after first sync complete")
		}
	})

	t.Run("syncTickMsg sends StatusUpdateMsg while waiting", func(t *testing.T) {
		t.Parallel()

		expectedStatus := &powersync.SyncStatus{
			Connected: false,
			Downloading: &powersync.DownloadProgress{
				Buckets: map[string]powersync.BucketProgress{
					"test": {SinceLast: 5, TargetCount: 10},
				},
			},
		}
		mock := &mockPowerSync{
			status:     powersync.StatusSyncing,
			syncStatus: expectedStatus,
			running:    true,
		}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = true
		syncer.firstSyncDone = false

		cmd := syncer.Update(syncTickMsg{})

		if cmd == nil {
			t.Fatal("expected batch command")
		}
		if !syncer.waiting {
			t.Error("should still be waiting")
		}
	})

	t.Run("syncTickMsg does nothing when not waiting", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{status: powersync.StatusConnected, running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = false

		cmd := syncer.Update(syncTickMsg{})

		if cmd != nil {
			t.Error("expected no command when not waiting")
		}
	})

	t.Run("SyncStatusQueryMsg returns SyncReadyMsg when ready", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{status: powersync.StatusConnected, running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = false

		cmd := syncer.Update(powersync.SyncStatusQueryMsg{})

		if cmd == nil {
			t.Fatal("expected command")
		}

		msg := cmd()
		if _, ok := msg.(powersync.SyncReadyMsg); !ok {
			t.Errorf("expected SyncReadyMsg, got %T", msg)
		}
	})

	t.Run("SyncStatusQueryMsg returns StatusUpdateMsg when waiting", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{status: powersync.StatusSyncing, running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = true

		cmd := syncer.Update(powersync.SyncStatusQueryMsg{})

		if cmd == nil {
			t.Fatal("expected command")
		}

		msg := cmd()
		statusMsg, ok := msg.(powersync.StatusUpdateMsg)
		if !ok {
			t.Fatalf("expected StatusUpdateMsg, got %T", msg)
		}
		if statusMsg.Status != powersync.StatusSyncing {
			t.Errorf("Status = %v, want %v", statusMsg.Status, powersync.StatusSyncing)
		}
	})
}

func TestSyncer_IsReady(t *testing.T) {
	t.Parallel()

	t.Run("false when sync is not running", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{running: false}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = false

		if syncer.IsReady() {
			t.Error("should not be ready when sync is not running")
		}
	})

	t.Run("false when still waiting for first sync", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = true

		if syncer.IsReady() {
			t.Error("should not be ready while waiting")
		}
	})

	t.Run("true when running and not waiting", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = false

		if !syncer.IsReady() {
			t.Error("should be ready when running and not waiting")
		}
	})
}

func TestSyncer_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stops the underlying sync", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{running: true}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))

		syncer.Stop()

		if !mock.stopped {
			t.Error("sync should be stopped")
		}
	})
}

func TestSyncer_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns syncReadyMsg on success", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))

		cmd := syncer.Start(nil, "account-123")

		if cmd == nil {
			t.Fatal("expected command")
		}

		msg := cmd()
		if _, ok := msg.(syncReadyMsg); !ok {
			t.Errorf("expected syncReadyMsg, got %T", msg)
		}
	})

	t.Run("returns StatusUpdateMsg with error on failure", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{
			startFunc: func(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error {
				return context.DeadlineExceeded
			},
		}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))

		cmd := syncer.Start(nil, "account-123")

		if cmd == nil {
			t.Fatal("expected command")
		}

		msg := cmd()
		statusMsg, ok := msg.(powersync.StatusUpdateMsg)
		if !ok {
			t.Fatalf("expected StatusUpdateMsg, got %T", msg)
		}
		if statusMsg.Status != powersync.StatusError {
			t.Errorf("Status = %v, want %v", statusMsg.Status, powersync.StatusError)
		}
		if statusMsg.LastError == nil {
			t.Error("LastError should not be nil")
		}
	})

	t.Run("sets waiting to true", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.waiting = false

		_ = syncer.Start(nil, "account-123")

		if !syncer.waiting {
			t.Error("should be waiting after Start")
		}
	})

	t.Run("resets firstSyncDone", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{}
		syncer := NewSyncer(context.Background(), nil, mock, logtest.New(t))
		syncer.firstSyncDone = true

		_ = syncer.Start(nil, "account-123")

		if syncer.firstSyncDone {
			t.Error("firstSyncDone should be reset")
		}
	})
}
