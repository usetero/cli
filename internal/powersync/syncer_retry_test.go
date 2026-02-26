package powersync_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/api/apitest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestSyncer_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("force-refreshes token on 401 and retries", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		var connectCalls atomic.Int32
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			n := connectCalls.Add(1)
			if n == 1 {
				return &api.Error{Kind: api.ErrorKindAuth, StatusCode: 401}
			}
			<-ctx.Done()
			return ctx.Err()
		}

		var forceRefreshed atomic.Bool
		refresher := &powersynctest.MockTokenRefresher{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				if forceRefreshed.Load() {
					return "new-token", nil
				}
				return "stale-token", nil
			},
			ForceRefreshAccessTokenFunc: func(ctx context.Context) (string, error) {
				forceRefreshed.Store(true)
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

		if !forceRefreshed.Load() {
			t.Error("expected ForceRefreshAccessToken to be called on 401")
		}
		if mock.Token != "new-token" {
			t.Errorf("Token = %q, want %q", mock.Token, "new-token")
		}
	})

	t.Run("auth errors are never fatal", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		var connectCalls atomic.Int32
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			n := connectCalls.Add(1)
			if n <= 5 {
				return &api.Error{Kind: api.ErrorKindAuth, StatusCode: 401}
			}
			<-ctx.Done()
			return ctx.Err()
		}

		refresher := &powersynctest.MockTokenRefresher{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
			ForceRefreshAccessTokenFunc: func(ctx context.Context) (string, error) {
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

		if _, ok := syncer.State().(*powersync.Error); ok {
			t.Error("auth errors should never be fatal")
		}
		if connectCalls.Load() <= 5 {
			t.Errorf("expected more than 5 connect calls, got %d", connectCalls.Load())
		}

		syncer.Stop()
	})

	t.Run("retries when token refresh fails", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)

		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			return &api.Error{Kind: api.ErrorKindAuth, StatusCode: 401}
		}

		var refreshCalls atomic.Int32
		refresher := &powersynctest.MockTokenRefresher{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
			ForceRefreshAccessTokenFunc: func(ctx context.Context) (string, error) {
				n := refreshCalls.Add(1)
				if n <= 2 {
					return "", fmt.Errorf("network error")
				}
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

		time.Sleep(4 * time.Second)

		if _, ok := syncer.State().(*powersync.Error); ok {
			t.Error("refresh failures should not be fatal")
		}
		if refreshCalls.Load() < 3 {
			t.Errorf("expected at least 3 refresh calls, got %d", refreshCalls.Load())
		}

		syncer.Stop()
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

		var connectCalls atomic.Int32
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			n := connectCalls.Add(1)
			if n == 1 {
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

		time.Sleep(2 * time.Second)
		syncer.Stop()

		if connectCalls.Load() < 2 {
			t.Errorf("expected at least 2 connect calls, got %d", connectCalls.Load())
		}

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

		var connectCalls atomic.Int32
		mock := apitest.NewMockClient()
		mock.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
			n := connectCalls.Add(1)
			if n == 1 {
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

		time.Sleep(2 * time.Second)
		syncer.Stop()

		if connectCalls.Load() < 2 {
			t.Errorf("expected at least 2 connect calls, got %d", connectCalls.Load())
		}
	})
}
