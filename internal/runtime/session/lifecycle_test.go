package session

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/runtime/session/sessiontest"
)

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
