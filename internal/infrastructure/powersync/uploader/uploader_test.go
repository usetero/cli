package uploader_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploadertest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestUploaderRunStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	raw := openTestDB(t)
	store := psdb.NewStore(raw)
	client := &uploadertest.Client{WriteCheckpoint: "1"}
	u, err := uploader.New(store, client, &uploadertest.TokenSource{Token: "tok"}, logtest.NewScope(t))
	if err != nil {
		t.Fatalf("uploader.New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = u.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v", err)
	}
	_, ok := <-u.Events()
	if ok {
		t.Fatal("events channel should be closed")
	}
}

func TestUploaderProcessesBatchAndCompletesCheckpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	insertCrud(t, raw, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{"content":"hello"}}`)

	store := psdb.NewStore(raw)
	client := &uploadertest.Client{WriteCheckpoint: "42"}
	handler := &countingHandler{}
	u, err := uploader.New(
		store,
		client,
		&uploadertest.TokenSource{Token: "tok"},
		logtest.NewScope(t),
		uploader.WithPolicy(uploader.RunPolicy{PollInterval: 5 * time.Millisecond, RetryDelay: 5 * time.Millisecond, MaxRetries: 1}),
		uploader.WithHandler(psdb.TableMessages, handler),
	)
	if err != nil {
		t.Fatalf("uploader.New() error = %v", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- u.Run(runCtx) }()

	event := waitEvent(t, u.Events(), 2*time.Second)
	if _, ok := event.(uploader.SyncingEvent); !ok {
		t.Fatalf("event = %T, want SyncingEvent", event)
	}
	cancel()
	<-done

	if handler.calls.Load() != 1 {
		t.Fatalf("handler calls = %d, want 1", handler.calls.Load())
	}
	if client.TokenSetCalls.Load() == 0 {
		t.Fatal("expected token to be set on client")
	}

	next, err := store.NextMutation(ctx)
	if err != nil {
		t.Fatalf("NextMutation() error = %v", err)
	}
	if next != nil {
		t.Fatalf("expected queue to be empty, got %+v", *next)
	}

	var targetOp int64
	if err := raw.QueryRow(ctx, "SELECT target_op FROM ps_buckets WHERE name = ?", string(psdb.LocalBucket)).Scan(&targetOp); err != nil {
		t.Fatalf("query target_op: %v", err)
	}
	if targetOp != 42 {
		t.Fatalf("target_op = %d, want 42", targetOp)
	}
}

func TestUploaderStalledThenRecovered(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	insertCrud(t, raw, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{"content":"hello"}}`)

	store := psdb.NewStore(raw)
	client := &uploadertest.Client{WriteCheckpoint: "11"}
	handler := &flakyHandler{failCount: 1}
	u, err := uploader.New(
		store,
		client,
		&uploadertest.TokenSource{Token: "tok"},
		logtest.NewScope(t),
		uploader.WithPolicy(uploader.RunPolicy{PollInterval: 5 * time.Millisecond, RetryDelay: 5 * time.Millisecond, MaxRetries: 0}),
		uploader.WithHandler(psdb.TableMessages, handler),
	)
	if err != nil {
		t.Fatalf("uploader.New() error = %v", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- u.Run(runCtx) }()

	var sawStalled, sawRecovered bool
	deadline := time.After(2 * time.Second)
	for !sawStalled || !sawRecovered {
		select {
		case ev := <-u.Events():
			switch ev.(type) {
			case uploader.StalledEvent:
				sawStalled = true
			case uploader.RecoveredEvent:
				sawRecovered = true
			}
		case <-deadline:
			t.Fatalf("timed out waiting stalled/recovered; stalled=%v recovered=%v", sawStalled, sawRecovered)
		}
	}

	cancel()
	<-done
}

func TestUploaderInvalidCheckpointStalls(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	insertCrud(t, raw, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{"content":"hello"}}`)

	store := psdb.NewStore(raw)
	client := &uploadertest.Client{WriteCheckpoint: "not-int"}
	u, err := uploader.New(
		store,
		client,
		&uploadertest.TokenSource{Token: "tok"},
		logtest.NewScope(t),
		uploader.WithPolicy(uploader.RunPolicy{PollInterval: 5 * time.Millisecond, RetryDelay: 5 * time.Millisecond, MaxRetries: 0}),
		uploader.WithHandler(psdb.TableMessages, &countingHandler{}),
	)
	if err != nil {
		t.Fatalf("uploader.New() error = %v", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- u.Run(runCtx) }()

	ev := waitEvent(t, u.Events(), 2*time.Second)
	stalled, ok := ev.(uploader.StalledEvent)
	if !ok {
		t.Fatalf("event = %T, want StalledEvent", ev)
	}
	if stalled.Error == nil || !strings.Contains(stalled.Error.Error(), "parse write checkpoint") {
		t.Fatalf("stalled error = %v", stalled.Error)
	}

	cancel()
	<-done
}

type countingHandler struct {
	calls atomic.Int32
}

func (h *countingHandler) Handle(context.Context, psdb.Mutation) error {
	h.calls.Add(1)
	return nil
}

type flakyHandler struct {
	calls     atomic.Int32
	failCount int32
}

func (h *flakyHandler) Handle(context.Context, psdb.Mutation) error {
	n := h.calls.Add(1)
	if n <= h.failCount {
		return errors.New("temporary failure")
	}
	return nil
}

func waitEvent(t *testing.T, ch <-chan uploader.Event, timeout time.Duration) uploader.Event {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(timeout):
		t.Fatal("timed out waiting for uploader event")
		return nil
	}
}

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	if err := extension.Register(); err != nil {
		t.Fatalf("extension.Register() error = %v", err)
	}
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "uploader-test.sqlite")
	db, err := sqlite.OpenBare(ctx, path)
	if err != nil {
		t.Fatalf("sqlite.OpenBare() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("extension.ApplySchema() error = %v", err)
	}
	return db
}

func seedLocalBucket(t *testing.T, db *sqlite.DB, last, target int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES (?, ?, ?)", string(psdb.LocalBucket), last, target); err != nil {
		t.Fatalf("seed local bucket: %v", err)
	}
}

func insertCrud(t *testing.T, db *sqlite.DB, id int64, txID *int64, data string) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.Exec(ctx, "INSERT INTO ps_crud (id, tx_id, data) VALUES (?, ?, ?)", id, txID, data); err != nil {
		t.Fatalf("insert crud row: %v", err)
	}
}
