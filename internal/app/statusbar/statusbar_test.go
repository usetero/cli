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
