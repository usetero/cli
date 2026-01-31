package loading_test

import (
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/loading"
)

type mockSyncState struct {
	db sqlite.Database
}

func (m *mockSyncState) DB() sqlite.Database { return m.db }

func TestLoading_IsComplete(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)
	org := api.Organization{ID: "org-1", Name: "Test Org"}
	account := api.Account{ID: "acc-1", Name: "Test Account"}

	t.Run("returns false when sync not complete", func(t *testing.T) {
		t.Parallel()

		syncState := &mockSyncState{db: nil}
		l := loading.New(theme, org, account, syncState)

		if l.IsComplete() {
			t.Error("expected IsComplete to return false when DB is nil")
		}
		if !l.IsBusy() {
			t.Error("expected IsBusy to return true when DB is nil")
		}
	})

	t.Run("returns true when sync complete", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		syncState := &mockSyncState{db: db}
		l := loading.New(theme, org, account, syncState)

		if !l.IsComplete() {
			t.Error("expected IsComplete to return true when DB is set")
		}
		if l.IsBusy() {
			t.Error("expected IsBusy to return false when DB is set")
		}
	})
}
