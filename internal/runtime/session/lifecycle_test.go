package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/runtime/session/sessiontest"
)

type gatedStorage struct {
	path        sqlite.DatabasePath
	acc2PathHit chan struct{}
}

func (s gatedStorage) DatabasePath(accountID sqlite.AccountID) (sqlite.DatabasePath, error) {
	if accountID == sqlite.AccountID("acc_2") {
		select {
		case s.acc2PathHit <- struct{}{}:
		default:
		}
	}
	return s.path, nil
}

func TestService_Switch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	storage := sessiontest.Storage{Path: sqlite.DatabasePath(dir + "/session.sqlite")}

	syncerOne := &sessiontest.Syncer{Ready: true}
	syncerTwo := &sessiontest.Syncer{Ready: true}
	syncers := []*sessiontest.Syncer{syncerOne, syncerTwo}
	syncerIdx := 0

	uploaderOne := sessiontest.NewUploader()
	uploaderTwo := sessiontest.NewUploader()
	uploaders := []*sessiontest.Uploader{uploaderOne, uploaderTwo}
	uploaderIdx := 0

	svc, err := NewService(
		storage,
		func() (Syncer, error) {
			v := syncers[syncerIdx]
			syncerIdx++
			return v, nil
		},
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			v := uploaders[uploaderIdx]
			uploaderIdx++
			return v, nil
		},
		logtest.NewScope(t),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	svc.setOpenDB(openBareDB(t))

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_1")); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := svc.Switch(context.Background(), tenancy.AccountID("acc_2")); err != nil {
		t.Fatalf("switch: %v", err)
	}
	defer svc.Stop()

	if !syncerOne.Stopped {
		t.Fatalf("expected previous syncer to stop")
	}
	state := svc.State()
	if !state.Running || state.AccountID != "acc_2" {
		t.Fatalf("unexpected state after switch: %+v", state)
	}
}

func TestService_StartValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewService(nil, nil, nil, logtest.NewScope(t)); err == nil {
		t.Fatalf("expected constructor validation error")
	}
}

func TestService_StartOpenDBError(t *testing.T) {
	t.Parallel()

	storage := sessiontest.Storage{Path: "/tmp/any.sqlite"}
	svc, err := NewService(
		storage,
		func() (Syncer, error) { return &sessiontest.Syncer{Ready: true}, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return sessiontest.NewUploader(), nil
		},
		logtest.NewScope(t),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	svc.setOpenDB(func(context.Context, sqlite.DatabasePath) (*sqlite.DB, error) {
		return nil, errors.New("boom")
	})

	if err := svc.Start(context.Background(), "acc_1"); err == nil {
		t.Fatalf("expected open db error")
	}
}

func TestService_ConcurrentStartWaitsForStop(t *testing.T) {
	t.Parallel()

	storage := gatedStorage{
		path:        sqlite.DatabasePath(t.TempDir() + "/session.sqlite"),
		acc2PathHit: make(chan struct{}, 1),
	}

	firstUploaderExit := make(chan struct{})
	firstUploaderCanceled := make(chan struct{}, 1)
	uploaderCalls := 0

	svc, err := NewService(
		storage,
		func() (Syncer, error) { return &sessiontest.Syncer{Ready: true}, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			uploaderCalls++
			uploader := sessiontest.NewUploader()
			if uploaderCalls == 1 {
				uploader.RunFn = func(ctx context.Context) error {
					<-ctx.Done()
					select {
					case firstUploaderCanceled <- struct{}{}:
					default:
					}
					<-firstUploaderExit
					return ctx.Err()
				}
			}
			return uploader, nil
		},
		logtest.NewScope(t),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	svc.setOpenDB(openBareDB(t))

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_1")); err != nil {
		t.Fatalf("start acc_1: %v", err)
	}

	stopDone := make(chan error, 1)
	go func() {
		stopDone <- svc.Stop()
	}()

	select {
	case <-firstUploaderCanceled:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for first uploader cancellation")
	}

	start2Done := make(chan error, 1)
	go func() {
		start2Done <- svc.Start(context.Background(), tenancy.AccountID("acc_2"))
	}()

	select {
	case <-storage.acc2PathHit:
		t.Fatalf("start acc_2 progressed before stop completed")
	case <-time.After(150 * time.Millisecond):
	}

	close(firstUploaderExit)

	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatalf("stop failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for stop completion")
	}

	select {
	case err := <-start2Done:
		if err != nil {
			t.Fatalf("start acc_2 failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for acc_2 start")
	}

	select {
	case <-storage.acc2PathHit:
	default:
		t.Fatalf("expected acc_2 database path resolution after stop completion")
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("final stop: %v", err)
	}
}
