package account

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploadertest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestNewStartsRuntimeAndExposesStatus(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	uploader := uploadertest.NewMock()

	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncer, uploader), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	status := runtime.Status()
	if !status.Running {
		t.Fatalf("expected running status")
	}
	if status.Scope.Account.ID != "acc_1" {
		t.Fatalf("unexpected account id: %s", status.Scope.Account.ID)
	}
	if runtime.DB() == nil {
		t.Fatalf("expected runtime db")
	}
}

func TestNewEmitsReadyWhenFirstSyncCompletes(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	syncer.StartFn = func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
		if onFirstSync != nil {
			onFirstSync()
		}
		return nil
	}

	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncer, uploadertest.NewMock()), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	assertEventSequence(t, runtime.Events(), EventStarting, EventReady)
}

func TestNewForwardsUploaderEvents(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	syncer.StartFn = func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
		return nil
	}
	uploader := uploadertest.NewMock()
	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncer, uploader), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	assertEventKind(t, runtime.Events(), EventStarting)

	uploader.EventsCh <- psuploader.SyncingEvent{ProcessedCount: 7}
	event := assertEventKind(t, runtime.Events(), EventSyncing)
	if event.ProcessedCount != 7 {
		t.Fatalf("unexpected processed count: %d", event.ProcessedCount)
	}
}

func TestNewCleansUpWhenSyncerStartFails(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	syncer.StartFn = func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
		return errors.New("boom")
	}

	_, err := newTestable(context.Background(), testScope(), testDeps(t, syncer, uploadertest.NewMock()), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err == nil {
		t.Fatalf("expected error")
	}
	if runtime := any(nil); runtime != nil {
		t.Fatalf("unexpected runtime: %v", runtime)
	}
}

func TestNewStopsSyncerWhenUploaderConstructionFails(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	stopped := false
	syncer.StopFn = func() { stopped = true }

	_, err := newTestable(context.Background(), testScope(), testableDeps{
		pathForScope: func(scope Scope) (sqlite.DatabasePath, error) {
			return sqlite.DatabasePath(filepath.Join(t.TempDir(), "tero.sqlite")), nil
		},
		openDB: func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error) {
			return sqlite.Open(ctx, path.String())
		},
		newSyncer: func() (testableSyncer, error) {
			return syncer, nil
		},
		newUploader: func(db *sqlite.DB, notifier interface {
			NotifyUploadCompleted(ctx context.Context) error
		}) (testableUploader, error) {
			return nil, errors.New("uploader build failed")
		},
	}, logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !stopped {
		t.Fatalf("expected syncer stop on uploader construction failure")
	}
}

func TestCloseStopsRuntimeAndUpdatesStatus(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	syncer.StartFn = func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
		return nil
	}
	stopped := false
	syncer.StopFn = func() { stopped = true }
	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncer, uploadertest.NewMock()), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	assertEventKind(t, runtime.Events(), EventStarting)

	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	if !stopped {
		t.Fatalf("expected syncer stop")
	}
	if runtime.DB() != nil {
		t.Fatalf("expected nil db after close")
	}
	if runtime.Status().Running {
		t.Fatalf("expected stopped status after close")
	}
	assertEventKind(t, runtime.Events(), EventStopped)
}

func TestStatusProjectsReadyState(t *testing.T) {
	t.Parallel()

	syncer := syncertest.NewMock()
	syncer.IsReadyFn = func() bool { return true }
	syncer.StateFn = func() pssyncer.State { return &pssyncer.Ready{} }
	uploader := uploadertest.NewMock()

	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncer, uploader), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	status := runtime.Status()
	if !status.Ready {
		t.Fatalf("expected ready status")
	}
	if _, ok := status.Sync.(*pssyncer.Ready); !ok {
		t.Fatalf("expected ready sync state, got %T", status.Sync)
	}
	if !status.HasCompletedInitialSync {
		t.Fatalf("expected initial sync marker to be persisted")
	}
	if !status.SessionReady {
		t.Fatalf("expected session ready status")
	}
}

func TestStatusSeparatesSessionReadyFromInitialSyncHistory(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "tero.sqlite")

	first := syncertest.NewMock()
	first.StartFn = func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
		if onFirstSync != nil {
			onFirstSync()
		}
		return nil
	}
	first.IsReadyFn = func() bool { return true }

	runtime, err := newTestable(
		context.Background(),
		testScope(),
		testDepsAtPath(path, first, uploadertest.NewMock()),
		logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)),
	)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status := runtime.Status()
	if !status.HasCompletedInitialSync {
		t.Fatalf("expected persisted initial sync marker after first run")
	}
	if !status.SessionReady {
		t.Fatalf("expected ready session on first run")
	}

	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close runtime: %v", err)
	}

	second := syncertest.NewMock()
	second.StartFn = func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
		return nil
	}
	second.IsReadyFn = func() bool { return false }
	second.StateFn = func() pssyncer.State { return &pssyncer.Syncing{} }

	runtime, err = newTestable(
		context.Background(),
		testScope(),
		testDepsAtPath(path, second, uploadertest.NewMock()),
		logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)),
	)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	status = runtime.Status()
	if !status.HasCompletedInitialSync {
		t.Fatalf("expected persisted initial sync marker on reopen")
	}
	if status.SessionReady {
		t.Fatalf("did not expect current session to be ready yet")
	}
}

func TestUploaderFatalErrorEmitsEventError(t *testing.T) {
	t.Parallel()

	uploader := uploadertest.NewMock()
	uploader.RunFn = func(ctx context.Context) error {
		return errors.New("fatal uploader failure")
	}

	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncertest.NewMock(), uploader), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	assertEventKind(t, runtime.Events(), EventStarting)
	event := assertEventKind(t, runtime.Events(), EventReady)
	if event.Kind != EventReady {
		t.Fatalf("unexpected event: %v", event.Kind)
	}
	errEvent := assertEventKind(t, runtime.Events(), EventError)
	if errEvent.Err == nil || !strings.Contains(errEvent.Err.Error(), "fatal uploader failure") {
		t.Fatalf("expected fatal uploader error, got %v", errEvent.Err)
	}
}

func TestUploaderCanceledDoesNotEmitEventError(t *testing.T) {
	t.Parallel()

	uploader := uploadertest.NewMock()
	uploader.RunFn = func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncertest.NewMock(), uploader), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	assertEventKind(t, runtime.Events(), EventStarting)
	assertEventKind(t, runtime.Events(), EventReady)

	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	assertEventKind(t, runtime.Events(), EventStopped)
	assertNoEvent(t, runtime.Events())
}

func TestCloseReturnsContextErrorWhenShutdownBlocks(t *testing.T) {
	t.Parallel()

	block := make(chan struct{})
	uploader := &blockingUploader{
		run: func(ctx context.Context) error {
			<-block
			return nil
		},
		events: make(chan psuploader.Event),
	}

	runtime, err := newTestable(context.Background(), testScope(), testDeps(t, syncertest.NewMock(), uploader), logging.RootScope(logging.NewWithWriter(nil, logging.LevelInfo)))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	assertEventKind(t, runtime.Events(), EventStarting)
	assertEventKind(t, runtime.Events(), EventReady)

	closeCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err = runtime.Close(closeCtx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	close(block)
}

func testScope() Scope {
	return Scope{
		Organization: tenancy.Organization{ID: "org_1", Name: "Org 1"},
		Account:      tenancy.Account{ID: "acc_1", Name: "Account 1"},
	}
}

func testDeps(t *testing.T, syncer testableSyncer, uploader testableUploader) testableDeps {
	t.Helper()

	return testDepsAtPath(filepath.Join(t.TempDir(), "tero.sqlite"), syncer, uploader)
}

func testDepsAtPath(path string, syncer testableSyncer, uploader testableUploader) testableDeps {
	return testableDeps{
		pathForScope: func(scope Scope) (sqlite.DatabasePath, error) {
			return sqlite.DatabasePath(path), nil
		},
		openDB: func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error) {
			return sqlite.Open(ctx, path.String())
		},
		newSyncer: func() (testableSyncer, error) {
			return syncer, nil
		},
		newUploader: func(db *sqlite.DB, notifier interface {
			NotifyUploadCompleted(ctx context.Context) error
		}) (testableUploader, error) {
			return uploader, nil
		},
	}
}

func assertEventKind(t *testing.T, events <-chan Event, kind EventKind) Event {
	t.Helper()

	select {
	case event := <-events:
		if event.Kind != kind {
			t.Fatalf("expected event kind %q, got %q", kind, event.Kind)
		}
		return event
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for event kind %q", kind)
		return Event{}
	}
}

func assertEventSequence(t *testing.T, events <-chan Event, kinds ...EventKind) {
	t.Helper()
	for _, kind := range kinds {
		assertEventKind(t, events, kind)
	}
}

func assertNoEvent(t *testing.T, events <-chan Event) {
	t.Helper()

	select {
	case event := <-events:
		t.Fatalf("expected no event, got %q", event.Kind)
	case <-time.After(25 * time.Millisecond):
	}
}

type blockingUploader struct {
	run    func(ctx context.Context) error
	events chan psuploader.Event
}

func (b *blockingUploader) Run(ctx context.Context) error {
	return b.run(ctx)
}

func (b *blockingUploader) Events() <-chan psuploader.Event {
	return b.events
}
