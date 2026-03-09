package session

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/runtime/session/sessiontest"
)

func TestServiceStart_ResetsDerivedDBOnSchemaApplyFailure(t *testing.T) {
	t.Parallel()

	path := sqlite.DatabasePath(t.TempDir() + "/session.sqlite")
	db, err := sqlite.OpenBare(context.Background(), path.String())
	if err != nil {
		t.Fatalf("seed sqlite db: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close seeded sqlite db: %v", err)
	}

	failing := &sessiontest.Syncer{StartErr: pssyncer.ErrApplySchema}
	succeeding := &sessiontest.Syncer{Ready: true}
	syncers := []Syncer{failing, succeeding}
	idx := 0

	uploader := sessiontest.NewUploader()
	svc := NewService(
		sessiontest.Storage{Path: path},
		func() (Syncer, error) {
			current := syncers[idx]
			idx++
			return current, nil
		},
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return uploader, nil
		},
		logtest.NewScope(t),
	)
	svc.setOpenDB(openBareDB(t))

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_1")); err != nil {
		t.Fatalf("start with schema reset retry: %v", err)
	}
	defer func() {
		if err := svc.Stop(); err != nil {
			t.Errorf("stop: %v", err)
		}
	}()

	if got := failing.StartCalls.Load(); got != 1 {
		t.Fatalf("expected first syncer start once, got %d", got)
	}
	if got := succeeding.StartCalls.Load(); got != 1 {
		t.Fatalf("expected retry syncer start once, got %d", got)
	}
	if _, err := os.Stat(path.String()); err != nil {
		t.Fatalf("expected recreated db file: %v", err)
	}
}

func TestResetSQLiteFiles_RemovesBaseAndSidecars(t *testing.T) {
	t.Parallel()

	path := sqlite.DatabasePath(t.TempDir() + "/session.sqlite")
	for _, target := range []string{path.String(), path.String() + "-wal", path.String() + "-shm"} {
		if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", target, err)
		}
	}

	if err := resetSQLiteFiles(path); err != nil {
		t.Fatalf("reset sqlite files: %v", err)
	}

	for _, target := range []string{path.String(), path.String() + "-wal", path.String() + "-shm"} {
		if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected %s to be removed, got %v", target, err)
		}
	}
}
