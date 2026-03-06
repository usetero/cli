package syncer_test

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestSyncerStartStop(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}}}
	cli := &syncertest.Client{
		SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, _ psclient.LineHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	s, err := syncer.New(
		"https://powersync.example",
		&syncertest.TokenSource{Token: "tok"},
		logtest.NewScope(t),
		syncer.WithControlFactory(func(_ *sqlite.DB) syncer.ControlPlane { return ctrl }),
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return cli }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        20 * time.Millisecond,
			ErrorStateAfter: 2,
		}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx, database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	s.Stop()

	if _, ok := s.State().(*syncer.Disconnected); !ok {
		t.Fatalf("State() = %T, want *Disconnected", s.State())
	}
}

func TestSyncerStartValidatesAccountID(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	s, err := syncer.New("https://powersync.example", &syncertest.TokenSource{Token: "tok"}, logtest.NewScope(t))
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}
	if err := s.Start(context.Background(), database, "", nil); err == nil {
		t.Fatal("expected error for empty accountID")
	}
}

func TestSyncerNotifyUploadCompletedBeforeStart(t *testing.T) {
	t.Parallel()

	s, err := syncer.New("https://powersync.example", &syncertest.TokenSource{Token: "tok"}, logtest.NewScope(t))
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}
	if err := s.NotifyUploadCompleted(context.Background()); err != nil {
		t.Fatalf("NotifyUploadCompleted() error = %v", err)
	}
}

func TestSyncerNotifyUploadCompletedAppliesInstructions(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{
		StartInstructions:        []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}},
		NotifyUploadInstructions: []extension.Instruction{{Type: extension.InstructionFetchCredentials}},
	}
	cli := &syncertest.Client{SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, _ psclient.LineHandler) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	tokens := &syncertest.TokenSource{Token: "tok-1"}

	s := newTestSyncer(t, tokens, ctrl, cli)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx, database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	tokens.Token = "tok-2"
	if err := s.NotifyUploadCompleted(context.Background()); err != nil {
		t.Fatalf("NotifyUploadCompleted() error = %v", err)
	}
	s.Stop()

	if cli.TokenValue != "tok-2" {
		t.Fatalf("client token = %q, want tok-2", cli.TokenValue)
	}
	if ctrl.NotifyTokenRefreshedCalls.Load() == 0 {
		t.Fatal("expected NotifyTokenRefreshed to be called")
	}
}

func TestSyncerAuthErrorForcesRefresh(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}}}
	var calls atomic.Int32
	cli := &syncertest.Client{SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, _ psclient.LineHandler) error {
		if calls.Add(1) == 1 {
			return &psclient.Error{Kind: psclient.ErrorKindAuth, StatusCode: 401, Message: "expired"}
		}
		<-ctx.Done()
		return ctx.Err()
	}}
	forceStarted := make(chan struct{})
	forceGate := make(chan struct{})
	tokens := &syncertest.TokenSource{
		Token:               "stale",
		ForceToken:          "fresh",
		ForceRefreshStarted: forceStarted,
		ForceRefreshGate:    forceGate,
	}

	s := newTestSyncer(t, tokens, ctrl, cli)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx, database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	select {
	case <-forceStarted:
	case <-time.After(time.Second):
		t.Fatal("force refresh did not start")
	}
	if _, ok := s.State().(*syncer.Reconnecting); !ok {
		t.Fatalf("State() = %T, want *Reconnecting", s.State())
	}
	close(forceGate)
	waitFor(t, time.Second, func() bool { return tokens.ForceCalls.Load() > 0 })
	s.Stop()

	if tokens.ForceCalls.Load() == 0 {
		t.Fatal("expected force refresh call")
	}
	if cli.TokenValue != "fresh" {
		t.Fatalf("client token = %q, want fresh", cli.TokenValue)
	}
}

func TestSyncerUpdateSyncStatusTracksProgress(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{
		StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}},
		SendTextInstructions: []extension.Instruction{{
			Type: extension.InstructionUpdateSyncStatus,
			SyncStatus: &extension.SyncStatus{
				Downloading: &extension.DownloadProgress{
					Buckets: map[string]extension.BucketProgress{
						"prio_3": {Priority: 3, SinceLast: 3, TargetCount: 10},
						"prio_4": {Priority: 4, SinceLast: 2, TargetCount: 5},
					},
				},
			},
		}},
	}
	cli := &syncertest.Client{SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
		if err := handler([]byte(`{"checkpoint":{"last_op_id":"1","buckets":[]}}`)); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}}

	s := newTestSyncer(t, &syncertest.TokenSource{Token: "tok"}, ctrl, cli)
	if err := s.Start(context.Background(), database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer s.Stop()

	waitFor(t, time.Second, func() bool {
		state, ok := s.State().(*syncer.Syncing)
		return ok && state.Progress != nil
	})

	state, ok := s.State().(*syncer.Syncing)
	if !ok || state.Progress == nil {
		t.Fatalf("State() = %#v, want Syncing with progress", s.State())
	}
	if state.Progress.Downloaded != 5 || state.Progress.Total != 15 {
		t.Fatalf("progress = %#v, want downloaded=5 total=15", state.Progress)
	}
}

func TestSyncerStartWhileRunningReturnsAlreadyStarted(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}}}
	cli := &syncertest.Client{
		SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, _ psclient.LineHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	s := newTestSyncer(t, &syncertest.TokenSource{Token: "tok"}, ctrl, cli)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx, database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("first Start() error = %v", err)
	}
	defer s.Stop()

	if err := s.Start(context.Background(), database, syncer.AccountID("acc-1"), nil); err != syncer.ErrAlreadyStarted {
		t.Fatalf("second Start() error = %v, want ErrAlreadyStarted", err)
	}
}

func TestSyncerPermanentErrorStops(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}}}
	cli := &syncertest.Client{SyncStreamFn: func(context.Context, *psclient.SyncStreamRequest, psclient.LineHandler) error {
		return &psclient.Error{Kind: psclient.ErrorKindPermanent, StatusCode: 400, Message: "bad request"}
	}}
	tokens := &syncertest.TokenSource{Token: "tok"}

	s := newTestSyncer(t, tokens, ctrl, cli)
	if err := s.Start(context.Background(), database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	waitFor(t, time.Second, func() bool {
		_, ok := s.State().(*syncer.Error)
		return ok
	})
	s.Stop()
}

func TestSyncerFirstSyncCallback(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{
		StartInstructions:    []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}},
		SendTextInstructions: []extension.Instruction{{Type: extension.InstructionDidCompleteSync}},
	}
	cli := &syncertest.Client{SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
		if err := handler([]byte(`{"checkpoint":{"last_op_id":"1","buckets":[]}}`)); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}}
	tokens := &syncertest.TokenSource{Token: "tok"}

	s := newTestSyncer(t, tokens, ctrl, cli)
	called := make(chan struct{}, 1)
	if err := s.Start(context.Background(), database, syncer.AccountID("acc-1"), func() { called <- struct{}{} }); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("first sync callback was not called")
	}
	s.Stop()
}

func TestSyncerFirstSyncCallbackOnlyOnce(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	ctrl := &syncertest.ControlPlane{
		StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}},
		SendTextInstructions: []extension.Instruction{
			{Type: extension.InstructionDidCompleteSync},
			{Type: extension.InstructionDidCompleteSync},
		},
	}
	cli := &syncertest.Client{SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
		if err := handler([]byte(`{"checkpoint":{"last_op_id":"1","buckets":[]}}`)); err != nil {
			return err
		}
		if err := handler([]byte(`{"checkpoint":{"last_op_id":"2","buckets":[]}}`)); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}}
	tokens := &syncertest.TokenSource{Token: "tok"}

	s := newTestSyncer(t, tokens, ctrl, cli)
	var called atomic.Int32
	if err := s.Start(context.Background(), database, syncer.AccountID("acc-1"), func() { called.Add(1) }); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	waitFor(t, time.Second, func() bool { return called.Load() >= 1 })
	time.Sleep(20 * time.Millisecond)
	s.Stop()

	if called.Load() != 1 {
		t.Fatalf("first sync callback calls = %d, want 1", called.Load())
	}
}

func TestSyncerSerializesControlPlaneCalls(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	blockSend := make(chan struct{})
	lineHandled := make(chan struct{})
	ctrl := &syncertest.ControlPlane{
		StartInstructions: []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &psclient.SyncStreamRequest{}}},
		SendTextLineHook: func() {
			close(lineHandled)
			<-blockSend
		},
	}
	cli := &syncertest.Client{SyncStreamFn: func(ctx context.Context, _ *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
		if err := handler([]byte(`{"checkpoint":{"last_op_id":"1","buckets":[]}}`)); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}}
	tokens := &syncertest.TokenSource{Token: "tok"}

	s := newTestSyncer(t, tokens, ctrl, cli)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx, database, syncer.AccountID("acc-1"), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer s.Stop()

	<-lineHandled // SendTextLine is active and blocked.

	notifyDone := make(chan error, 1)
	go func() {
		notifyDone <- s.NotifyUploadCompleted(context.Background())
	}()

	time.Sleep(30 * time.Millisecond)
	if ctrl.MaxConcurrentCalls.Load() != 1 {
		t.Fatalf("max concurrent control calls = %d, want 1", ctrl.MaxConcurrentCalls.Load())
	}

	close(blockSend)
	if err := <-notifyDone; err != nil {
		t.Fatalf("NotifyUploadCompleted() error = %v", err)
	}
}

func newTestSyncer(t *testing.T, tokens *syncertest.TokenSource, ctrl *syncertest.ControlPlane, cli *syncertest.Client) *syncer.Syncer {
	t.Helper()
	s, err := syncer.New(
		"https://powersync.example",
		tokens,
		logtest.NewScope(t),
		syncer.WithControlFactory(func(_ *sqlite.DB) syncer.ControlPlane { return ctrl }),
		syncer.WithClientFactory(func(_ syncer.Endpoint) syncer.Client { return cli }),
		syncer.WithRetryPolicy(syncer.RetryPolicy{
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        20 * time.Millisecond,
			ErrorStateAfter: 2,
		}),
	)
	if err != nil {
		t.Fatalf("syncer.New() error = %v", err)
	}
	return s
}

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	if err := extension.Register(); err != nil {
		t.Fatalf("extension.Register() error = %v", err)
	}
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "syncer.sqlite")
	database, err := sqlite.OpenBare(ctx, path)
	if err != nil {
		t.Fatalf("sqlite.OpenBare() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func waitFor(t *testing.T, timeout time.Duration, predicate func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
