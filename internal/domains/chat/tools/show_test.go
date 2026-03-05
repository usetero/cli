package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestShowTool_Run(t *testing.T) {
	t.Parallel()

	sqlite.SetExtensionPath("")
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "show.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(context.Background(), "CREATE TABLE things (id TEXT PRIMARY KEY, name TEXT NOT NULL)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(context.Background(), "INSERT INTO things (id, name) VALUES ('thing_1', 'Thing')"); err != nil {
		t.Fatalf("insert row: %v", err)
	}

	tool := NewShowTool(db)
	out, err := tool.Run(context.Background(), json.RawMessage(`{"entity":"thing","sql":"select id from things where id='thing_1'"}`))
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	var got ShowResult
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Entity != "thing" || got.ID != "thing_1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestShowTool_RequiresIDOrSQL(t *testing.T) {
	t.Parallel()

	sqlite.SetExtensionPath("")
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "show.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	tool := NewShowTool(db)
	_, err = tool.Run(context.Background(), json.RawMessage(`{"entity":"thing"}`))
	if err == nil || !strings.Contains(err.Error(), "either id or sql is required") {
		t.Fatalf("expected missing id/sql error, got %v", err)
	}
}
