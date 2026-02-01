package database

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

func TestDatabase_New(t *testing.T) {
	t.Parallel()

	t.Run("creates database model", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		config := &powersync.Config{Endpoint: "http://localhost:8084"}
		logger := logtest.New(t)

		db := New(ctx, config, nil, nil, nil, logger)

		if db == nil {
			t.Fatal("expected non-nil database")
		}
		if db.ctx != ctx {
			t.Error("context not set")
		}
		if db.powersyncConfig != config {
			t.Error("config not set")
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
			db: &mockDatabase{},
		}

		if d.IsReady() {
			t.Error("should not be ready when syncer is nil")
		}
	})

	t.Run("false when syncer is not ready", func(t *testing.T) {
		t.Parallel()

		d := &Database{
			db: &mockDatabase{},
			syncer: &Syncer{
				waiting: true,
			},
		}

		if d.IsReady() {
			t.Error("should not be ready when syncer is not ready")
		}
	})

	t.Run("true when db and syncer are ready", func(t *testing.T) {
		t.Parallel()

		d := &Database{
			db: &mockDatabase{},
			syncer: &Syncer{
				syncer:  &mockSyncer{},
				waiting: false,
			},
		}

		if !d.IsReady() {
			t.Error("should be ready")
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

		mock := &mockDatabase{}
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

		mock := &mockSyncer{running: true}
		mockDB := &mockDatabase{}

		d := &Database{
			db: mockDB,
			syncer: &Syncer{
				syncer: mock,
			},
		}

		d.Close()

		if mock.running {
			t.Error("syncer should be stopped")
		}
		if !mockDB.closed {
			t.Error("db should be closed")
		}
	})
}

// mockDatabase implements sqlite.Database for testing.
type mockDatabase struct {
	closed bool
}

func (m *mockDatabase) Messages() sqlite.Messages           { return nil }
func (m *mockDatabase) Conversations() sqlite.Conversations { return nil }
func (m *mockDatabase) Subscribe() *sqlite.Subscription     { return nil }
func (m *mockDatabase) Close() error                        { m.closed = true; return nil }
func (m *mockDatabase) DB() *sqlite.DB                      { return nil }

var _ sqlite.Database = (*mockDatabase)(nil)
