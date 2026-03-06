package integration_test

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

func TestReplay_InvariantCheckpointFixtureDoesNotFatal(t *testing.T) {
	t.Parallel()

	raw := openPowerSyncTestDB(t)
	fixturePath := filepath.Join("..", "extension", "testdata", "checkpoint_lines.ndjson")
	client := &replayStreamClient{path: fixturePath}
	scope := logging.RootScope(logging.NewWithWriter(io.Discard, logging.LevelInfo))

	s, err := syncer.New(
		"https://powersync.example",
		&syncerTokenSource{token: "tok"},
		scope,
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return client }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{InitialDelay: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond, ErrorStateAfter: 2}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := s.Start(ctx, raw, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("syncer.Start() error = %v", err)
	}
	waitUntil(t, time.Second, func() bool { return client.lineCount.Load() > 0 })
	s.Stop()
	cancel()

	if client.calls.Load() == 0 {
		t.Fatal("expected replay stream to connect")
	}
	if _, ok := s.State().(*syncer.Error); ok {
		t.Fatalf("syncer entered fatal error state: %#v", s.State())
	}
}

func TestReplay_InvariantSanitizedFixtureDoesNotFatal(t *testing.T) {
	t.Parallel()

	raw := openPowerSyncTestDB(t)
	fixturePath := filepath.Join("..", "extension", "testdata", "dev-sanitized.ndjson")
	client := &replayStreamClient{path: fixturePath}
	scope := logging.RootScope(logging.NewWithWriter(io.Discard, logging.LevelInfo))

	s, err := syncer.New(
		"https://powersync.example",
		&syncerTokenSource{token: "tok"},
		scope,
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return client }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{InitialDelay: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond, ErrorStateAfter: 2}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := s.Start(ctx, raw, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("syncer.Start() error = %v", err)
	}
	waitUntil(t, time.Second, func() bool { return client.lineCount.Load() > 0 })
	s.Stop()
	cancel()

	if client.calls.Load() == 0 {
		t.Fatal("expected replay stream to connect")
	}
	if _, ok := s.State().(*syncer.Error); ok {
		t.Fatalf("syncer entered fatal error state: %#v", s.State())
	}
}
