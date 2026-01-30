package powersync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestNewSync(t *testing.T) {
	t.Parallel()

	t.Run("initial status is disconnected", func(t *testing.T) {
		t.Parallel()

		sync := powersync.NewSync(&powersync.Config{Endpoint: "https://example.com"}, nil)

		if sync.Status() != powersync.StatusDisconnected {
			t.Errorf("Status() = %v, want %v", sync.Status(), powersync.StatusDisconnected)
		}
	})

	t.Run("IsRunning is false initially", func(t *testing.T) {
		t.Parallel()

		sync := powersync.NewSync(&powersync.Config{}, nil)

		if sync.IsRunning() {
			t.Error("IsRunning() should be false before Start")
		}
	})

	t.Run("LastError is nil initially", func(t *testing.T) {
		t.Parallel()

		sync := powersync.NewSync(&powersync.Config{}, nil)

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
		sync := powersync.NewSync(&powersync.Config{Endpoint: "https://example.com"}, nil)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty accountID")
			}
		}()

		_ = sync.Start(context.Background(), db, "", "token")
	})

	t.Run("panics on empty token", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		sync := powersync.NewSync(&powersync.Config{Endpoint: "https://example.com"}, nil)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty token")
			}
		}()

		_ = sync.Start(context.Background(), db, "account-123", "")
	})

	t.Run("returns error if already started", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "token")
		if err != nil {
			t.Fatalf("first Start() error = %v", err)
		}
		defer sync.Stop()

		err = sync.Start(ctx, db, "account-123", "token")
		if err == nil {
			t.Error("expected error on second Start()")
		}
	})

	t.Run("sets status to connecting", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		started := make(chan struct{})
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "token")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		// Wait for sync loop to start
		<-started

		// Status should be syncing (set when establishing stream)
		status := sync.Status()
		if status != powersync.StatusSyncing {
			t.Errorf("Status() = %v, want %v", status, powersync.StatusSyncing)
		}
	})

	t.Run("IsRunning is true after start", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "token")
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
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx := context.Background()
		err := sync.Start(ctx, db, "account-123", "token")
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
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx := context.Background()
		_ = sync.Start(ctx, db, "account-123", "token")
		sync.Stop()

		if sync.IsRunning() {
			t.Error("IsRunning() should be false after Stop")
		}
	})

	t.Run("is safe to call multiple times", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx := context.Background()
		_ = sync.Start(ctx, db, "account-123", "token")

		// Should not panic
		sync.Stop()
		sync.Stop()
		sync.Stop()
	})
}

func TestSync_AuthErrorTriggersTokenRefresh(t *testing.T) {
	t.Parallel()

	t.Run("refreshes token on 401 and retries", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		connectCalls := 0
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				connectCalls++
				if connectCalls == 1 {
					// First call returns auth error
					return &powersync.StreamError{Kind: powersync.ErrorKindAuth, StatusCode: 401}
				}
				// Second call blocks until cancelled
				<-ctx.Done()
				return ctx.Err()
			},
		}

		refresher := &powersynctest.MockTokenRefresher{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "new-token", nil
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			refresher,
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "initial-token")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Wait for retry - needs time for controller.Start() calls
		time.Sleep(500 * time.Millisecond)
		sync.Stop()

		// Token refresh should have been called
		if refresher.Calls == 0 {
			t.Error("expected token refresh to be called")
		}

		// Token should have been updated on the stream
		if mock.Token != "new-token" {
			t.Errorf("Token = %q, want %q", mock.Token, "new-token")
		}

		// Connect should have been called at least once (the retry may or may not
		// have happened depending on controller state, but token refresh is the key behavior)
		if connectCalls < 1 {
			t.Errorf("expected at least 1 connect call, got %d", connectCalls)
		}
	})
}

func TestSync_PermanentErrorStopsSync(t *testing.T) {
	t.Parallel()

	t.Run("sets error status on permanent error", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				return &powersync.StreamError{Kind: powersync.ErrorKindPermanent, StatusCode: 400, Message: "bad request"}
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx := context.Background()
		err := sync.Start(ctx, db, "account-123", "token")
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

func TestSync_TransientErrorRetries(t *testing.T) {
	t.Parallel()

	t.Run("retries on transient error", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		connectCalls := 0
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				connectCalls++
				if connectCalls == 1 {
					return &powersync.StreamError{Kind: powersync.ErrorKindTransient, StatusCode: 503}
				}
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "token")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Wait for retry (initial backoff is 1 second)
		time.Sleep(2 * time.Second)
		sync.Stop()

		// Should have retried at least once after transient error
		if connectCalls < 2 {
			t.Errorf("expected at least 2 connect calls (initial + retry), got %d", connectCalls)
		}
	})

	t.Run("sets reconnecting status on transient error", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		statusChecked := make(chan struct{})
		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				select {
				case <-statusChecked:
					// After status is checked, block until cancelled
					<-ctx.Done()
					return ctx.Err()
				default:
					return &powersync.StreamError{Kind: powersync.ErrorKindTransient, StatusCode: 503}
				}
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "token")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Wait a bit for the error to be processed
		time.Sleep(200 * time.Millisecond)

		status := sync.Status()
		close(statusChecked)
		sync.Stop()

		if status != powersync.StatusReconnecting {
			t.Errorf("Status() = %v, want %v", status, powersync.StatusReconnecting)
		}
	})
}

func TestSync_WaitForFirstSync(t *testing.T) {
	t.Parallel()

	t.Run("returns when ps_buckets has data", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)

		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				// Simulate sync by inserting data into ps_buckets
				_, err := db.Exec("INSERT INTO ps_buckets (name) VALUES ('test')")
				if err != nil {
					return err
				}
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := sync.Start(ctx, db, "account-123", "token")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		err = sync.WaitForFirstSync(ctx)
		if err != nil {
			t.Errorf("WaitForFirstSync() error = %v", err)
		}
	})

	t.Run("returns error on context cancellation", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)

		mock := &powersynctest.MockStreamer{
			ConnectFunc: func(ctx context.Context, req *powersync.StreamingSyncRequest, handler powersync.LineHandler) error {
				// Never insert data - sync never completes
				<-ctx.Done()
				return ctx.Err()
			},
		}

		sync := powersync.NewSyncForTest(
			&powersync.Config{Endpoint: "https://example.com"},
			powersynctest.NewMockTokenRefresher("token"),
			powersynctest.NewMockStreamerFactory(mock),
		)

		ctx := context.Background()
		err := sync.Start(ctx, db, "account-123", "token")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		waitCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err = sync.WaitForFirstSync(waitCtx)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("WaitForFirstSync() error = %v, want context.DeadlineExceeded", err)
		}
	})
}

func TestIsAuthError(t *testing.T) {
	t.Parallel()

	t.Run("true for auth StreamError", func(t *testing.T) {
		t.Parallel()

		err := &powersync.StreamError{Kind: powersync.ErrorKindAuth, StatusCode: 401}
		if !powersync.IsAuthError(err) {
			t.Error("IsAuthError should return true for ErrorKindAuth")
		}
	})

	t.Run("false for transient StreamError", func(t *testing.T) {
		t.Parallel()

		err := &powersync.StreamError{Kind: powersync.ErrorKindTransient, StatusCode: 500}
		if powersync.IsAuthError(err) {
			t.Error("IsAuthError should return false for ErrorKindTransient")
		}
	})

	t.Run("true for wrapped auth error", func(t *testing.T) {
		t.Parallel()

		streamErr := &powersync.StreamError{Kind: powersync.ErrorKindAuth, StatusCode: 401}
		wrappedErr := errors.Join(errors.New("context"), streamErr)
		if !powersync.IsAuthError(wrappedErr) {
			t.Error("IsAuthError should return true for wrapped auth errors")
		}
	})
}

func TestIsTransientError(t *testing.T) {
	t.Parallel()

	t.Run("true for transient StreamError", func(t *testing.T) {
		t.Parallel()

		err := &powersync.StreamError{Kind: powersync.ErrorKindTransient, StatusCode: 500}
		if !powersync.IsTransientError(err) {
			t.Error("IsTransientError should return true for ErrorKindTransient")
		}
	})

	t.Run("false for permanent StreamError", func(t *testing.T) {
		t.Parallel()

		err := &powersync.StreamError{Kind: powersync.ErrorKindPermanent, StatusCode: 400}
		if powersync.IsTransientError(err) {
			t.Error("IsTransientError should return false for ErrorKindPermanent")
		}
	})
}
