package session

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploadertest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/runtime/session/sessiontest"
)

func TestService_UsesPowerSyncTestkits(t *testing.T) {
	t.Parallel()

	storage := sessiontest.Storage{Path: sqlite.DatabasePath(t.TempDir() + "/session.sqlite")}
	syncer := &syncertest.Mock{
		StartFn: func(_ context.Context, _ *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
			if accountID != "acc_kit" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if onFirstSync != nil {
				onFirstSync()
			}
			return nil
		},
		IsReadyFn: func() bool { return true },
	}
	uploader := uploadertest.NewMock()

	svc, err := NewService(
		storage,
		func() (Syncer, error) { return syncer, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return uploader, nil
		},
		logtest.NewScope(t),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	svc.setOpenDB(openBareDB(t))

	if err := svc.Start(context.Background(), tenancy.AccountID("acc_kit")); err != nil {
		t.Fatalf("start: %v", err)
	}
	if !svc.IsReady() {
		t.Fatalf("expected ready from syncertest mock")
	}
	if err := svc.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
}
