package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/auth/authtest"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/upload"
)

type testStorage struct {
	dbPath string
}

func (s testStorage) DatabasePath(accountID string) (string, error) {
	return s.dbPath, nil
}

func (s testStorage) ClearDatabase(accountID string) error {
	return nil
}

func (s testStorage) Clear() error {
	return nil
}

type testUploader struct{}

func (testUploader) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (testUploader) Events() <-chan upload.Event {
	return nil
}

func TestOpenDatabase_ClosesPreviousDatabase(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	storage := testStorage{dbPath: filepath.Join(tmp, "next.sqlite")}
	prev := sqlitetest.NewMockDB()

	m := &Model{
		ctx:     ctx,
		scope:   logtest.NewScope(t),
		storage: storage,
		db:      prev,
	}

	if err := m.openDatabase("acc_123"); err != nil {
		t.Fatalf("openDatabase() error = %v", err)
	}

	if !prev.Closed {
		t.Fatalf("expected previous database to be closed")
	}
	if m.db == nil {
		t.Fatalf("expected new database to be set")
	}
	if m.db == prev {
		t.Fatalf("expected new database instance")
	}

	if err := m.db.Close(); err != nil {
		t.Fatalf("close new database: %v", err)
	}
}

func TestShutdown_CleansRuntimeResources(t *testing.T) {
	db := sqlitetest.NewMockDB()
	syncer := powersynctest.NewMockSyncer()
	syncerStopped := false
	syncer.StopFunc = func() {
		syncerStopped = true
	}
	cancelled := false

	m := &Model{
		scope:         logtest.NewScope(t),
		syncer:        syncer,
		sessionCancel: func() { cancelled = true },
		db:            db,
		uploader:      testUploader{},
	}

	m.shutdown()

	if !cancelled {
		t.Fatalf("expected session cancel to be called")
	}
	if !syncerStopped {
		t.Fatalf("expected syncer stop to be called")
	}
	if !db.Closed {
		t.Fatalf("expected db to be closed")
	}
	if m.sessionCancel != nil {
		t.Fatalf("expected sessionCancel to be cleared")
	}
	if m.db != nil {
		t.Fatalf("expected db to be nil")
	}
	if m.uploader != nil {
		t.Fatalf("expected uploader to be nil")
	}
}

func TestStartSync_RequiresOpenDatabase(t *testing.T) {
	m := &Model{}

	if err := m.startSync("acc_123"); err == nil {
		t.Fatalf("expected error when db is not open")
	}
}

func TestStartSync_InitializesSessionAndUploader(t *testing.T) {
	scope := logtest.NewScope(t)
	db := sqlitetest.OpenBareDB(t)

	syncer := powersynctest.NewMockSyncer()
	startCalled := false
	syncer.StartFunc = func(ctx context.Context, gotDB sqlite.DB, accountID string, onFirstSync func()) error {
		startCalled = true
		if gotDB != db {
			t.Fatalf("syncer received unexpected db instance")
		}
		if accountID != "acc_123" {
			t.Fatalf("syncer received accountID=%q", accountID)
		}
		return nil
	}

	mockClient := apitest.NewMockClient()
	var scopedAccountID domain.AccountID
	mockClient.SetAccountIDFunc = func(accountID domain.AccountID) {
		scopedAccountID = accountID
	}

	authService := &authtest.MockAuth{
		GetAccessTokenFunc: func(ctx context.Context) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
	}

	m := &Model{
		ctx:         context.Background(),
		scope:       scope,
		cfg:         &config.CLIConfig{PowerSyncEndpoint: "https://powersync.example.com"},
		db:          db,
		syncer:      syncer,
		services:    graphql.NewAPIServices(mockClient, scope),
		authService: authService,
	}

	if err := m.startSync("acc_123"); err != nil {
		t.Fatalf("startSync() error = %v", err)
	}
	t.Cleanup(m.shutdown)

	if !startCalled {
		t.Fatalf("expected syncer start to be called")
	}
	if m.sessionCancel == nil {
		t.Fatalf("expected session cancel to be initialized")
	}
	if m.uploader == nil {
		t.Fatalf("expected uploader to be initialized")
	}
	if scopedAccountID != domain.AccountID("acc_123") {
		t.Fatalf("expected services account scope to be set, got %q", scopedAccountID)
	}
}
