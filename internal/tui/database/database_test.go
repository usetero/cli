package database

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestDatabase_New(t *testing.T) {
	t.Parallel()

	t.Run("creates database model with dependencies", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockSync := &mockPowerSync{}
		logger := logtest.New(t)

		db := New(ctx, nil, mockSync, powersynctest.NewMockClient(), powersynctest.NewMockTokenRefresher("token"), nil, nil, logger)

		if db == nil {
			t.Fatal("expected non-nil database")
		}
	})
}

func TestDatabase_IsReady(t *testing.T) {
	t.Parallel()

	t.Run("false when db is nil", func(t *testing.T) {
		t.Parallel()

		d := &Database{}

		if d.IsReady() {
			t.Error("should not be ready when db is nil")
		}
	})

	t.Run("false when syncer is nil", func(t *testing.T) {
		t.Parallel()

		d := &Database{
			db: sqlitetest.NewMockDB(),
		}

		if d.IsReady() {
			t.Error("should not be ready when syncer is nil")
		}
	})

	t.Run("false when syncer is not ready", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{running: false}
		d := &Database{
			db: sqlitetest.NewMockDB(),
			syncer: &Syncer{
				sync:    mock,
				waiting: true,
			},
		}

		if d.IsReady() {
			t.Error("should not be ready when syncer is not ready")
		}
	})

	t.Run("true when db and syncer are ready", func(t *testing.T) {
		t.Parallel()

		mock := &mockPowerSync{running: true}
		d := &Database{
			db: sqlitetest.NewMockDB(),
			syncer: &Syncer{
				sync:    mock,
				waiting: false,
			},
		}

		if !d.IsReady() {
			t.Error("should be ready when db exists and syncer is ready")
		}
	})
}

func TestDatabase_DB(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when not opened", func(t *testing.T) {
		t.Parallel()

		d := &Database{}

		if d.DB() != nil {
			t.Error("expected nil db")
		}
	})

	t.Run("returns db when opened", func(t *testing.T) {
		t.Parallel()

		mock := sqlitetest.NewMockDB()
		d := &Database{db: mock}

		if d.DB() != mock {
			t.Error("expected mock db")
		}
	})
}

func TestDatabase_Close(t *testing.T) {
	t.Parallel()

	t.Run("safe to call when nothing initialized", func(t *testing.T) {
		t.Parallel()

		d := &Database{}

		// Should not panic
		d.Close()
	})

	t.Run("stops syncer and closes db", func(t *testing.T) {
		t.Parallel()

		mockSync := &mockPowerSync{running: true}
		mockDB := sqlitetest.NewMockDB()

		d := &Database{
			db: mockDB,
			syncer: &Syncer{
				sync: mockSync,
			},
		}

		d.Close()

		if !mockSync.stopped {
			t.Error("syncer should be stopped")
		}
		if !mockDB.Closed {
			t.Error("db should be closed")
		}
	})
}

func TestDatabase_Update(t *testing.T) {
	t.Parallel()

	t.Run("databaseOpenedMsg initializes syncer and uploader", func(t *testing.T) {
		t.Parallel()

		mockSync := &mockPowerSync{}
		d := &Database{
			ctx:             context.Background(),
			syncClient:      mockSync,
			powersyncClient: powersynctest.NewMockClient(),
			tokenRefresher:  powersynctest.NewMockTokenRefresher("token"),
			logger:          logtest.New(t),
		}

		mockDB := sqlitetest.NewMockDB()
		cmd := d.Update(databaseOpenedMsg{db: mockDB, accountID: "acc-123"})

		if d.db != mockDB {
			t.Error("db should be set")
		}
		if d.syncer == nil {
			t.Error("syncer should be initialized")
		}
		if d.uploader == nil {
			t.Error("uploader should be initialized")
		}
		if cmd == nil {
			t.Error("should return batch command to start syncer and uploader")
		}
	})

	t.Run("delegates to syncer when initialized", func(t *testing.T) {
		t.Parallel()

		mockSync := &mockPowerSync{status: powersync.StatusConnected, running: true}
		d := &Database{
			syncer: &Syncer{
				sync:    mockSync,
				waiting: false,
			},
		}

		// SyncStatusQueryMsg should be delegated to syncer
		cmd := d.Update(powersync.SyncStatusQueryMsg{})

		if cmd == nil {
			t.Error("should delegate to syncer and return command")
		}
	})
}
