package database

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
)

// mockSyncer implements powersync.Syncer for testing.
type mockSyncer struct {
	status     powersync.Status
	syncStatus *powersync.SyncStatus
	lastError  error
	running    bool
}

func (m *mockSyncer) Stop()                             { m.running = false }
func (m *mockSyncer) Status() powersync.Status          { return m.status }
func (m *mockSyncer) SyncStatus() *powersync.SyncStatus { return m.syncStatus }
func (m *mockSyncer) LastError() error                  { return m.lastError }
func (m *mockSyncer) IsRunning() bool                   { return m.running }

func TestSyncer_Update(t *testing.T) {
	t.Parallel()

	t.Run("syncReadyMsg starts polling", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{status: powersync.StatusSyncing, running: true}

		cmd := syncer.Update(syncReadyMsg{syncer: mock})

		if cmd == nil {
			t.Fatal("expected tick command")
		}
		if syncer.syncer != mock {
			t.Error("syncer not set")
		}
		if !syncer.waiting {
			t.Error("should be waiting")
		}
	})

	t.Run("syncTickMsg sends SyncReadyMsg when firstSyncDone", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{status: powersync.StatusSyncing, running: true}
		syncer.syncer = mock
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
			t.Error("should not be waiting after first sync")
		}
	})

	t.Run("syncTickMsg sends StatusUpdateMsg while waiting", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{
			status: powersync.StatusSyncing,
			syncStatus: &powersync.SyncStatus{
				Connected: false,
				Downloading: &powersync.DownloadProgress{
					Buckets: map[string]powersync.BucketProgress{
						"test": {SinceLast: 5, TargetCount: 10},
					},
				},
			},
			running: true,
		}
		syncer.syncer = mock
		syncer.waiting = true
		syncer.firstSyncDone = false

		cmd := syncer.Update(syncTickMsg{})

		if cmd == nil {
			t.Fatal("expected batch command")
		}

		// The batch contains status update + tick
		// We can't easily inspect batch, but verify state is unchanged
		if !syncer.waiting {
			t.Error("should still be waiting")
		}
	})

	t.Run("syncTickMsg does nothing when not waiting", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{status: powersync.StatusSyncing}
		syncer.syncer = mock
		syncer.waiting = false

		cmd := syncer.Update(syncTickMsg{})

		if cmd != nil {
			t.Error("expected no command when not waiting")
		}
	})

	t.Run("syncTickMsg does nothing when syncer is nil", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		syncer.waiting = true

		cmd := syncer.Update(syncTickMsg{})

		if cmd != nil {
			t.Error("expected no command when syncer is nil")
		}
	})

	t.Run("SyncStatusQueryMsg returns SyncReadyMsg when ready", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{status: powersync.StatusConnected}
		syncer.syncer = mock
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

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{status: powersync.StatusSyncing}
		syncer.syncer = mock
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

	t.Run("false when syncer is nil", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))

		if syncer.IsReady() {
			t.Error("should not be ready when syncer is nil")
		}
	})

	t.Run("false when waiting", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		syncer.syncer = &mockSyncer{}
		syncer.waiting = true

		if syncer.IsReady() {
			t.Error("should not be ready while waiting")
		}
	})

	t.Run("true when syncer exists and not waiting", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		syncer.syncer = &mockSyncer{}
		syncer.waiting = false

		if !syncer.IsReady() {
			t.Error("should be ready")
		}
	})
}

func TestSyncer_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stops the syncer", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))
		mock := &mockSyncer{running: true}
		syncer.syncer = mock

		syncer.Stop()

		if mock.running {
			t.Error("syncer should be stopped")
		}
	})

	t.Run("safe to call when syncer is nil", func(t *testing.T) {
		t.Parallel()

		syncer := NewSyncer(context.Background(), nil, nil, logtest.New(t))

		// Should not panic
		syncer.Stop()
	})
}
