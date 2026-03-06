package integration_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

func TestInvariant_PipelineDrainsQueueAndAdvancesCheckpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openPowerSyncTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	insertCrud(t, raw, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{"content":"hello"}}`)
	store := psdb.NewStore(raw)

	client := &pipelineClient{checkpoint: "7", streamLine: []byte(`{"checkpoint":{"last_op_id":"1","buckets":[]}}`)}
	s, err := syncer.New(
		"https://powersync.example",
		&syncerTokenSource{token: "tok"},
		logtest.NewScope(t),
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return client }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{InitialDelay: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond, ErrorStateAfter: 2}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}

	syncCtx, syncCancel := context.WithCancel(ctx)
	defer syncCancel()
	if err := s.Start(syncCtx, raw, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("syncer.Start() error = %v", err)
	}
	defer s.Stop()

	handler := &countingMutationHandler{}
	u := uploader.New(
		store,
		client,
		&uploaderTokenSource{token: "tok"},
		logtest.NewScope(t),
		uploader.WithPolicy(uploader.RunPolicy{PollInterval: 5 * time.Millisecond, RetryDelay: 5 * time.Millisecond, MaxRetries: 1}),
		uploader.WithHandler(psdb.TableMessages, handler),
		uploader.WithSyncNotifier(s),
	)

	uCtx, uCancel := context.WithCancel(ctx)
	uDone := make(chan error, 1)
	go func() { uDone <- u.Run(uCtx) }()

	ev := waitUploaderEvent(t, u.Events(), 2*time.Second)
	if _, ok := ev.(uploader.SyncingEvent); !ok {
		t.Fatalf("event = %T, want uploader.SyncingEvent", ev)
	}
	uCancel()
	if err := <-uDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("uploader.Run() error = %v", err)
	}

	if handler.calls.Load() != 1 {
		t.Fatalf("handler calls = %d, want 1", handler.calls.Load())
	}
	if client.writeCheckpointCalls.Load() == 0 {
		t.Fatal("expected write-checkpoint request")
	}

	next, err := store.NextMutation(ctx)
	if err != nil {
		t.Fatalf("store.NextMutation() error = %v", err)
	}
	if next != nil {
		t.Fatalf("expected queue empty, got %+v", *next)
	}

	var targetOp int64
	if err := raw.QueryRow(ctx, "SELECT target_op FROM ps_buckets WHERE name = ?", string(psdb.LocalBucket)).Scan(&targetOp); err != nil {
		t.Fatalf("query target_op: %v", err)
	}
	if targetOp != 7 {
		t.Fatalf("target_op = %d, want 7", targetOp)
	}

	if _, ok := s.State().(*syncer.Error); ok {
		t.Fatalf("syncer entered fatal error state: %#v", s.State())
	}
}

func TestInvariant_GoldenReplayPipelineLeavesConsistentSQLiteState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openPowerSyncTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	insertCrud(t, raw, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{"content":"hello"}}`)
	store := psdb.NewStore(raw)

	fixturePath := filepath.Join("..", "extension", "testdata", "dev-sanitized.ndjson")
	client := &replayStreamClient{
		path:       fixturePath,
		checkpoint: "170",
	}

	s, err := syncer.New(
		"https://powersync.example",
		&syncerTokenSource{token: "tok"},
		logtest.NewScope(t),
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return client }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{InitialDelay: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond, ErrorStateAfter: 2}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}

	syncCtx, syncCancel := context.WithCancel(ctx)
	defer syncCancel()
	if err := s.Start(syncCtx, raw, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("syncer.Start() error = %v", err)
	}
	defer s.Stop()

	waitUntil(t, time.Second, func() bool { return client.lineCount.Load() > 0 })

	handler := &countingMutationHandler{}
	u := uploader.New(
		store,
		client,
		&uploaderTokenSource{token: "tok"},
		logtest.NewScope(t),
		uploader.WithPolicy(uploader.RunPolicy{PollInterval: 5 * time.Millisecond, RetryDelay: 5 * time.Millisecond, MaxRetries: 1}),
		uploader.WithHandler(psdb.TableMessages, handler),
		uploader.WithSyncNotifier(s),
	)

	uCtx, uCancel := context.WithCancel(ctx)
	uDone := make(chan error, 1)
	go func() { uDone <- u.Run(uCtx) }()

	ev := waitUploaderEvent(t, u.Events(), 2*time.Second)
	if _, ok := ev.(uploader.SyncingEvent); !ok {
		t.Fatalf("event = %T, want uploader.SyncingEvent", ev)
	}
	uCancel()
	if err := <-uDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("uploader.Run() error = %v", err)
	}

	if handler.calls.Load() != 1 {
		t.Fatalf("handler calls = %d, want 1", handler.calls.Load())
	}
	if client.calls.Load() == 0 || client.lineCount.Load() == 0 {
		t.Fatalf("expected replay client to process fixture, calls=%d lines=%d", client.calls.Load(), client.lineCount.Load())
	}

	next, err := store.NextMutation(ctx)
	if err != nil {
		t.Fatalf("store.NextMutation() error = %v", err)
	}
	if next != nil {
		t.Fatalf("expected queue empty, got %+v", *next)
	}

	var targetOp int64
	if err := raw.QueryRow(ctx, "SELECT target_op FROM ps_buckets WHERE name = ?", string(psdb.LocalBucket)).Scan(&targetOp); err != nil {
		t.Fatalf("query target_op: %v", err)
	}
	if targetOp != 170 {
		t.Fatalf("target_op = %d, want 170", targetOp)
	}

	if _, ok := s.State().(*syncer.Error); ok {
		t.Fatalf("syncer entered fatal error state: %#v", s.State())
	}
}
