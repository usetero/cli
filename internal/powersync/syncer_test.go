package powersync_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/api/apitest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

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

	t.Run("panics on empty accountID", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		syncer := powersync.NewSyncer(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			logtest.NewScope(t),
		)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty accountID")
			}
		}()

		_ = syncer.Start(context.Background(), db, "", nil)
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
			// Success
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

		// Should not panic
		syncer.Stop()
		syncer.Stop()
		syncer.Stop()
	})
}

func TestSyncer_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("refreshes token on 401 and retries", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		connectCalls := 0
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			connectCalls++
			if connectCalls == 1 {
				return &api.Error{Kind: api.ErrorKindAuth, StatusCode: 401}
			}
			<-ctx.Done()
			return ctx.Err()
		}

		refresher := &powersynctest.MockTokenRefresher{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "new-token", nil
			},
		}

		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			refresher,
			mock,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		time.Sleep(500 * time.Millisecond)
		syncer.Stop()

		if refresher.Calls == 0 {
			t.Error("expected token refresh to be called")
		}
		if mock.Token != "new-token" {
			t.Errorf("Token = %q, want %q", mock.Token, "new-token")
		}
	})

	t.Run("transitions to error state on permanent error", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			return &api.Error{Kind: api.ErrorKindPermanent, StatusCode: 400, Message: "bad request"}
		}

		syncer := powersynctest.NewSyncerWithMockClient(
			"https://example.com",
			powersynctest.NewMockTokenRefresher("token"),
			mock,
		)

		ctx := context.Background()
		err := syncer.Start(ctx, db, "account-123", nil)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		errState, ok := syncer.State().(*powersync.Error)
		if !ok {
			t.Fatalf("State() = %T, want *Error", syncer.State())
		}
		if errState.Err == nil {
			t.Error("Error.Err should not be nil")
		}
	})

	t.Run("retries on non-API error with backoff", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		connectCalls := 0
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			connectCalls++
			if connectCalls == 1 {
				// Simulate an extension error (not an api.Error)
				return fmt.Errorf("powersync_control: invalid state: No iteration is active")
			}
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

		// Wait for retry (initial backoff is 1 second)
		time.Sleep(2 * time.Second)
		syncer.Stop()

		if connectCalls < 2 {
			t.Errorf("expected at least 2 connect calls, got %d", connectCalls)
		}

		// Should NOT be in error state — should have recovered
		if _, ok := syncer.State().(*powersync.Error); ok {
			t.Error("non-API errors should be retried, not fatal")
		}
		if _, ok := syncer.State().(*powersync.Reconnecting); ok {
			t.Error("expected recovery, not reconnecting")
		}
	})

	t.Run("marks degraded after repeated failures", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			return fmt.Errorf("powersync_control: invalid state: No iteration is active")
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

		// Wait long enough for errorStateAfter (3) retries with backoff (1s + 2s + 4s)
		time.Sleep(8 * time.Second)

		state, ok := syncer.State().(*powersync.Reconnecting)
		if !ok {
			t.Fatalf("State() = %T, want *Reconnecting", syncer.State())
		}
		if !state.Degraded {
			t.Error("expected Degraded to be true after repeated failures")
		}

		syncer.Stop()
	})

	t.Run("retries on transient error with backoff", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		connectCalls := 0
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			connectCalls++
			if connectCalls == 1 {
				return &api.Error{Kind: api.ErrorKindTransient, StatusCode: 503}
			}
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

		// Wait for retry (initial backoff is 1 second)
		time.Sleep(2 * time.Second)
		syncer.Stop()

		if connectCalls < 2 {
			t.Errorf("expected at least 2 connect calls, got %d", connectCalls)
		}
	})
}
