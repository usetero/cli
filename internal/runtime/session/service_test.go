package session

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/runtime/session/sessiontest"
)

func openBareDB(t *testing.T) dbOpenFunc {
	t.Helper()
	return func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error) {
		return sqlite.OpenBare(ctx, path.String())
	}
}

func TestService_StartAndStop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	storage := sessiontest.Storage{Path: sqlite.DatabasePath(dir + "/session.sqlite")}
	syncer := &sessiontest.Syncer{Ready: true}
	uploader := sessiontest.NewUploader()

	svc := NewService(
		storage,
		func() (Syncer, error) { return syncer, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return uploader, nil
		},
		logtest.NewScope(t),
	)
	svc.setOpenDB(openBareDB(t))

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_1")); err != nil {
		t.Fatalf("start: %v", err)
	}

	state := svc.State()
	if !state.Running {
		t.Fatalf("expected running state")
	}
	if state.AccountID != "acc_1" {
		t.Fatalf("unexpected account id: %s", state.AccountID)
	}
	if svc.DB() == nil {
		t.Fatalf("expected open db")
	}
	if !svc.IsReady() {
		t.Fatalf("expected ready")
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if svc.State().Running {
		t.Fatalf("expected stopped state")
	}
	if svc.DB() != nil {
		t.Fatalf("expected nil db after stop")
	}
	if !syncer.Stopped {
		t.Fatalf("expected syncer stop")
	}
}

func TestService_ForwardUploaderEvents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	storage := sessiontest.Storage{Path: sqlite.DatabasePath(dir + "/session.sqlite")}
	syncer := &sessiontest.Syncer{Ready: true}
	uploader := sessiontest.NewUploader()

	svc := NewService(
		storage,
		func() (Syncer, error) { return syncer, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return uploader, nil
		},
		logtest.NewScope(t),
	)
	svc.setOpenDB(openBareDB(t))

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_1")); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() {
		if stopErr := svc.Stop(); stopErr != nil {
			t.Errorf("stop: %v", stopErr)
		}
	}()

	uploader.EventsCh <- psuploader.SyncingEvent{ProcessedCount: 2}
	uploader.EventsCh <- psuploader.StalledEvent{Error: context.DeadlineExceeded, Table: psdb.TableName("messages"), RowID: "row_1"}
	uploader.EventsCh <- psuploader.RecoveredEvent{}

	gotSyncing := false
	gotStalled := false
	gotRecovered := false
	deadline := time.After(2 * time.Second)
	for !gotSyncing || !gotStalled || !gotRecovered {
		select {
		case ev := <-svc.Events():
			switch ev.Kind {
			case EventStarting, EventReady, EventStopped, EventError:
				// Not relevant for this test; uploader mapping assertions below.
			case EventSyncing:
				gotSyncing = true
			case EventStalled:
				gotStalled = true
			case EventRecovered:
				gotRecovered = true
			}
		case <-deadline:
			t.Fatalf("timed out waiting for mapped events")
		}
	}
}

func TestService_SyncState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	storage := sessiontest.Storage{Path: sqlite.DatabasePath(dir + "/session.sqlite")}
	syncer := &sessiontest.Syncer{Ready: false, StateValue: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 3, Total: 10}}}
	uploader := sessiontest.NewUploader()

	svc := NewService(
		storage,
		func() (Syncer, error) { return syncer, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return uploader, nil
		},
		logtest.NewScope(t),
	)
	svc.setOpenDB(openBareDB(t))

	if _, ok := svc.SyncState().(*pssyncer.Disconnected); !ok {
		t.Fatalf("expected disconnected state before start")
	}

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_1")); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() {
		if stopErr := svc.Stop(); stopErr != nil {
			t.Errorf("stop: %v", stopErr)
		}
	}()

	state := svc.SyncState()
	syncing, ok := state.(*pssyncer.Syncing)
	if !ok {
		t.Fatalf("expected syncing state, got %T", state)
	}
	if syncing.Progress == nil || syncing.Progress.Downloaded != 3 || syncing.Progress.Total != 10 {
		t.Fatalf("unexpected progress: %+v", syncing.Progress)
	}
}
