package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

func TestRecovery_StalledThenRecoveredAndDrained(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openPowerSyncTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	insertCrud(t, raw, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{"content":"hello"}}`)
	store := psdb.NewStore(raw)

	client := &pipelineClient{checkpoint: "9"}
	handler := &flakyMutationHandler{failFirstN: 1}
	u, err := uploader.New(
		store,
		client,
		&uploaderTokenSource{token: "tok"},
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
	deadline := time.After(3 * time.Second)
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
			t.Fatalf("timed out waiting for stalled/recovered, stalled=%v recovered=%v", sawStalled, sawRecovered)
		}
	}

	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("uploader.Run() error = %v", err)
	}

	next, err := store.NextMutation(ctx)
	if err != nil {
		t.Fatalf("store.NextMutation() error = %v", err)
	}
	if next != nil {
		t.Fatalf("expected queue drained after recovery, got %+v", *next)
	}
}
