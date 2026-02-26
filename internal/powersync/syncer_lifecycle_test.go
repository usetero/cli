package powersync_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/api/apitest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

type closeCountCapture struct {
	closeCalls int
}

func (c *closeCountCapture) CaptureLine([]byte) {}
func (c *closeCountCapture) Close() error {
	c.closeCalls++
	return nil
}

func TestNewSyncer(t *testing.T) {
	t.Parallel()

	t.Run("initial state is disconnected", func(t *testing.T) {
		t.Parallel()

		syncer := powersync.NewSyncer("https://example.com", nil, logtest.NewScope(t))

		if _, ok := syncer.State().(*powersync.Disconnected); !ok {
			t.Errorf("State() = %T, want *Disconnected", syncer.State())
		}
	})

	t.Run("IsReady is false initially", func(t *testing.T) {
		t.Parallel()

		syncer := powersync.NewSyncer("https://example.com", nil, logtest.NewScope(t))

		if syncer.IsReady() {
			t.Error("IsReady() should be false before Start")
		}
	})
}

func TestSyncer_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns error on empty accountID", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersync.NewSyncer(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			logtest.NewScope(t),
		)

		err := syncer.Start(context.Background(), db, "", nil)
		if err == nil {
			t.Error("expected error for empty accountID")
		}
	})

	t.Run("returns error if already started", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			apitest.NewMockClient(),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("first Start() error = %v", err)
		}
		defer syncer.Stop()

		err = syncer.Start(ctx, db, "account-123", nil)
		if err == nil {
			t.Error("expected error on second Start()")
		}
	})

	t.Run("transitions to syncing when stream connects", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		started := make(chan struct{})
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		}

		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer syncer.Stop()

		<-started

		state := syncer.State()
		_, isSyncing := state.(*powersync.Syncing)
		_, isConnecting := state.(*powersync.Connecting)
		if !isSyncing && !isConnecting {
			t.Errorf("State() = %T, want *Syncing or *Connecting", state)
		}
	})

	t.Run("processes lines from stream", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		handlerCalled := make(chan struct{})
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			if err := handler([]byte(`{"token_expires_in":3600}`)); err != nil {
				return err
			}
			close(handlerCalled)
			<-ctx.Done()
			return ctx.Err()
		}

		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer syncer.Stop()

		select {
		case <-handlerCalled:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for stream line to be processed")
		}
	})

	t.Run("does not panic with nil onFirstSync callback", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			apitest.NewMockClient(),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		syncer.Stop()
	})
}

func TestSyncer_Stop(t *testing.T) {
	t.Parallel()

	t.Run("transitions to disconnected", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			apitest.NewMockClient(),
		)

		ctx := context.Background()
		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		syncer.Stop()

		if _, ok := syncer.State().(*powersync.Disconnected); !ok {
			t.Errorf("State() = %T, want *Disconnected", syncer.State())
		}
	})

	t.Run("IsReady is false after stop", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			apitest.NewMockClient(),
		)

		ctx := context.Background()
		_ = syncer.Start(ctx, db, "account-123", nil)
		syncer.Stop()

		if syncer.IsReady() {
			t.Error("IsReady() should be false after Stop")
		}
	})

	t.Run("is safe to call multiple times", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			apitest.NewMockClient(),
		)

		ctx := context.Background()
		_ = syncer.Start(ctx, db, "account-123", nil)

		syncer.Stop()
		syncer.Stop()
		syncer.Stop()
	})

	t.Run("closes configured stream capture only once", func(t *testing.T) {
		t.Parallel()

		capture := &closeCountCapture{}
		syncer := powersync.NewSyncer(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			logtest.NewScope(t),
			powersync.WithStreamCapture(capture),
		)

		syncer.Stop()
		syncer.Stop()

		if capture.closeCalls != 1 {
			t.Fatalf("capture close calls = %d, want 1", capture.closeCalls)
		}
	})
}

func TestSyncer_NotifyUploadCompleted(t *testing.T) {
	t.Parallel()

	t.Run("is safe before start", func(t *testing.T) {
		t.Parallel()

		syncer := powersync.NewSyncer(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			logtest.NewScope(t),
		)

		if err := syncer.NotifyUploadCompleted(context.Background()); err != nil {
			t.Fatalf("NotifyUploadCompleted() error = %v", err)
		}
	})

	t.Run("is safe after stop", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			apitest.NewMockClient(),
		)

		if err := syncer.Start(context.Background(), db, "account-123", nil); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		syncer.Stop()

		if err := syncer.NotifyUploadCompleted(context.Background()); err != nil {
			t.Fatalf("NotifyUploadCompleted() error = %v", err)
		}
	})
}
