package powersync_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestNewSync(t *testing.T) {
	t.Parallel()

	t.Run("initial status is disconnected", func(t *testing.T) {
		t.Parallel()

		sync := powersync.NewSync("https://example.com", nil, logtest.New(t))

		if sync.Status() != powersync.StatusDisconnected {
			t.Errorf("Status() = %v, want %v", sync.Status(), powersync.StatusDisconnected)
		}
	})

	t.Run("IsRunning is false initially", func(t *testing.T) {
		t.Parallel()

		sync := powersync.NewSync("https://example.com", nil, logtest.New(t))

		if sync.IsRunning() {
			t.Error("IsRunning() should be false before Start")
		}
	})

	t.Run("LastError is nil initially", func(t *testing.T) {
		t.Parallel()

		sync := powersync.NewSync("https://example.com", nil, logtest.New(t))

		if sync.LastError() != nil {
			t.Error("LastError() should be nil initially")
		}
	})
}

func TestSync_Start(t *testing.T) {
	t.Parallel()

	t.Run("panics on empty accountID", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersync.NewSync("https://example.com", powersynctest.NewMockTokenRefresher("token"), logtest.New(t))

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty accountID")
			}
		}()

		_ = sync.Start(context.Background(), db, "", nil)
	})

	t.Run("returns error if already started", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockClient(),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("first Start() error = %v", err)
		}
		defer sync.Stop()

		err = sync.Start(ctx, db, "account-123", nil)
		if err == nil {
			t.Error("expected error on second Start()")
		}
	})

	t.Run("sets status to syncing when stream connects", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		started := make(chan struct{})
		mock := powersynctest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		}

		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		<-started

		if sync.Status() != powersync.StatusSyncing {
			t.Errorf("Status() = %v, want %v", sync.Status(), powersync.StatusSyncing)
		}
	})

	t.Run("IsRunning is true after start", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockClient(),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		if !sync.IsRunning() {
			t.Error("IsRunning() should be true after Start")
		}
	})
}

func TestSync_Stop(t *testing.T) {
	t.Parallel()

	t.Run("sets status to disconnected", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockClient(),
		)

		ctx := context.Background()
		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		sync.Stop()

		if sync.Status() != powersync.StatusDisconnected {
			t.Errorf("Status() = %v, want %v", sync.Status(), powersync.StatusDisconnected)
		}
	})

	t.Run("IsRunning is false after stop", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockClient(),
		)

		ctx := context.Background()
		_ = sync.Start(ctx, db, "account-123", nil)
		sync.Stop()

		if sync.IsRunning() {
			t.Error("IsRunning() should be false after Stop")
		}
	})

	t.Run("is safe to call multiple times", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockClient(),
		)

		ctx := context.Background()
		_ = sync.Start(ctx, db, "account-123", nil)

		// Should not panic
		sync.Stop()
		sync.Stop()
		sync.Stop()
	})
}

func TestSync_OnFirstSync(t *testing.T) {
	t.Parallel()

	t.Run("does not panic with nil callback", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockClient(),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		sync.Stop()
	})
}

func TestSync_AuthError(t *testing.T) {
	t.Parallel()

	t.Run("refreshes token on 401 and retries", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		connectCalls := 0
		mock := powersynctest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
			connectCalls++
			if connectCalls == 1 {
				return &powersync.ClientError{Kind: powersync.ErrorKindAuth, StatusCode: 401}
			}
			<-ctx.Done()
			return ctx.Err()
		}

		refresher := &powersynctest.MockTokenRefresher{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "new-token", nil
			},
		}

		sync := powersync.NewSync("https://example.com", refresher, logtest.New(t))
		sync.SetClientFactory(powersynctest.NewMockClientFactory(mock))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Wait for retry
		time.Sleep(500 * time.Millisecond)
		sync.Stop()

		if refresher.Calls == 0 {
			t.Error("expected token refresh to be called")
		}

		if mock.Token != "new-token" {
			t.Errorf("Token = %q, want %q", mock.Token, "new-token")
		}
	})
}

func TestSync_PermanentError(t *testing.T) {
	t.Parallel()

	t.Run("sets error status", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		mock := powersynctest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
			return &powersync.ClientError{Kind: powersync.ErrorKindPermanent, StatusCode: 400, Message: "bad request"}
		}

		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx := context.Background()
		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Wait for sync loop to process error
		time.Sleep(100 * time.Millisecond)

		if sync.Status() != powersync.StatusError {
			t.Errorf("Status() = %v, want %v", sync.Status(), powersync.StatusError)
		}
		if sync.LastError() == nil {
			t.Error("LastError() should not be nil")
		}
	})
}

func TestSync_TransientError(t *testing.T) {
	t.Parallel()

	t.Run("retries after backoff", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		connectCalls := 0
		mock := powersynctest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
			connectCalls++
			if connectCalls == 1 {
				return &powersync.ClientError{Kind: powersync.ErrorKindTransient, StatusCode: 503}
			}
			<-ctx.Done()
			return ctx.Err()
		}

		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Wait for retry (initial backoff is 1 second)
		time.Sleep(2 * time.Second)
		sync.Stop()

		if connectCalls < 2 {
			t.Errorf("expected at least 2 connect calls, got %d", connectCalls)
		}
	})

	t.Run("sets reconnecting status", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		statusChecked := make(chan struct{})
		mock := powersynctest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *powersync.SyncStreamRequest, handler powersync.LineHandler) error {
			select {
			case <-statusChecked:
				<-ctx.Done()
				return ctx.Err()
			default:
				return &powersync.ClientError{Kind: powersync.ErrorKindTransient, StatusCode: 503}
			}
		}

		sync := powersynctest.NewSyncWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		time.Sleep(200 * time.Millisecond)

		status := sync.Status()
		close(statusChecked)
		sync.Stop()

		if status != powersync.StatusReconnecting {
			t.Errorf("Status() = %v, want %v", status, powersync.StatusReconnecting)
		}
	})
}
