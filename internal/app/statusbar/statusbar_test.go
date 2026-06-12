package statusbar

import (
	"context"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
)

func TestFetchWorkspaceCountAndUpdate(t *testing.T) {
	db := sqlitetest.OpenBareDB(t)
	ctx := context.Background()

	if _, err := db.Exec(ctx, "CREATE TABLE workspaces (id TEXT PRIMARY KEY)"); err != nil {
		t.Fatalf("create workspaces table: %v", err)
	}
	if _, err := db.Exec(ctx, "INSERT INTO workspaces (id) VALUES (?), (?)", "w1", "w2"); err != nil {
		t.Fatalf("insert workspaces: %v", err)
	}

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	msg := m.fetchWorkspaceCount(db)()
	countMsg, ok := msg.(workspaceCountLoadedMsg)
	if !ok {
		t.Fatalf("expected workspaceCountLoadedMsg, got %T", msg)
	}
	if countMsg.err != nil {
		t.Fatalf("unexpected fetch error: %v", countMsg.err)
	}
	if countMsg.count != 2 {
		t.Fatalf("expected count=2, got %d", countMsg.count)
	}

	m.org = "Acme"
	m.workspace = "Prod"
	m.Update(countMsg)

	view := m.renderOrgWorkspace()
	if !strings.Contains(view, "Acme / Prod") {
		t.Fatalf("expected org/workspace view, got %q", view)
	}
}

func TestFetchWorkspaceCountReturnsErrorWhenTableMissing(t *testing.T) {
	db := sqlitetest.OpenBareDB(t)
	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")

	msg := m.fetchWorkspaceCount(db)()
	countMsg, ok := msg.(workspaceCountLoadedMsg)
	if !ok {
		t.Fatalf("expected workspaceCountLoadedMsg, got %T", msg)
	}
	if countMsg.err == nil {
		t.Fatalf("expected error when workspaces table is missing")
	}
}

func TestSyncStatusStaysWiredOutsideDrawerTabs(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")

	// syncStatus renders the brand sync dot but is no longer a drawer tab, so
	// the lifecycle must reach it independently of m.tabs. Clearing the tabs
	// isolates that wiring: with the syncer present, Init must still start the
	// sync poll loop.
	m.tabs = nil
	if m.Init() == nil {
		t.Fatal("Init() must start syncStatus polling even with no drawer tabs")
	}
}

func TestBuildTabsMirrorsProductSurfaces(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")

	want := []struct {
		group string
		label string
	}{
		{group: "Control Plane", label: "Issues"},
		{group: "Control Plane", label: "Checks"},
		{group: "Data Plane", label: "Services"},
		{group: "Data Plane", label: "Log events"},
		{group: "Data Plane", label: "Edge instances"},
	}

	if len(m.tabs) != len(want) {
		t.Fatalf("tab count = %d, want %d", len(m.tabs), len(want))
	}
	for i, tab := range m.tabs {
		if tab.Label() != want[i].label {
			t.Fatalf("tab %d label = %q, want %q", i, tab.Label(), want[i].label)
		}
		grouped, ok := tab.(groupedDrawerTab)
		if !ok {
			t.Fatalf("tab %d does not expose a group", i)
		}
		if grouped.GroupLabel() != want[i].group {
			t.Fatalf("tab %d group = %q, want %q", i, grouped.GroupLabel(), want[i].group)
		}
	}
}
