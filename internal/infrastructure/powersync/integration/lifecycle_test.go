package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

func TestInvariant_LifecycleStartStopRestartIsSafe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	raw := openPowerSyncTestDB(t)
	seedLocalBucket(t, raw, 0, 0)
	store := psdb.NewStore(raw)

	client := &pipelineClient{checkpoint: "1"}
	s, err := syncer.New(
		"https://powersync.example",
		&syncerTokenSource{token: "tok"},
		logtest.NewScope(t),
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return client }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{InitialDelay: 10 * time.Millisecond, MaxDelay: 30 * time.Millisecond, ErrorStateAfter: 2}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}

	start := func() context.CancelFunc {
		runCtx, cancel := context.WithCancel(ctx)
		if err := s.Start(runCtx, raw, syncer.AccountID("acc-1"), nil); err != nil {
			t.Fatalf("syncer.Start() error = %v", err)
		}
		return cancel
	}

	cancel1 := start()
	time.Sleep(20 * time.Millisecond)
	cancel1()
	s.Stop()

	if _, ok := s.State().(*syncer.Disconnected); !ok {
		t.Fatalf("state after first stop = %T", s.State())
	}

	cancel2 := start()
	time.Sleep(20 * time.Millisecond)
	cancel2()
	s.Stop()

	u := uploader.New(
		store,
		client,
		&uploaderTokenSource{token: "tok"},
		logtest.NewScope(t),
		uploader.WithPolicy(uploader.RunPolicy{PollInterval: 5 * time.Millisecond, RetryDelay: 5 * time.Millisecond, MaxRetries: 0}),
	)

	uCtx, uCancel := context.WithCancel(ctx)
	uDone := make(chan error, 1)
	go func() { uDone <- u.Run(uCtx) }()
	time.Sleep(20 * time.Millisecond)
	uCancel()
	if err := <-uDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("uploader.Run() error = %v", err)
	}
}
