package syncstatus

import (
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
)

func TestPendingPollDoesNotOverlapFetches(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com")
	db := sqlitetest.OpenBareDB(t)
	m.SetDB(db)

	m.pendingFetch = true
	if cmd := m.Update(pendingUploadsPollTickMsg{}); cmd == nil {
		t.Fatalf("expected pending poll to keep polling even while fetch is in-flight")
	}
	if !m.pendingFetch {
		t.Fatalf("expected in-flight flag to remain set")
	}
}

func TestPendingUploadsLoadedMsgClearsInFlightFlag(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com")
	m.pendingFetch = true

	m.Update(pendingUploadsLoadedMsg{total: 42})

	if m.pendingFetch {
		t.Fatalf("expected pending message to clear in-flight flag")
	}
	if m.totalPending != 42 {
		t.Fatalf("expected pending total to update, got %d", m.totalPending)
	}
}
