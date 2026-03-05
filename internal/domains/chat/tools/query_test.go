package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestQueryTool_Run(t *testing.T) {
	t.Parallel()

	sqlite.SetExtensionPath("")
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "query.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(context.Background(), "CREATE TABLE items (id TEXT PRIMARY KEY, name TEXT NOT NULL)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(context.Background(), "INSERT INTO items (id, name) VALUES ('1', 'alpha'), ('2', 'beta')"); err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	tool := NewQueryTool(db)
	out, err := tool.Run(context.Background(), json.RawMessage(`{"sql":"SELECT id, name FROM items ORDER BY id"}`))
	if err != nil {
		t.Fatalf("run query: %v", err)
	}

	var got QueryResult
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got.Rows))
	}
}

func TestQueryTool_RejectsMutations(t *testing.T) {
	t.Parallel()

	sqlite.SetExtensionPath("")
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "query.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	tool := NewQueryTool(db)
	_, err = tool.Run(context.Background(), json.RawMessage(`{"sql":"DELETE FROM items"}`))
	if err == nil || !strings.Contains(err.Error(), "read-only SELECT") {
		t.Fatalf("expected read-only rejection, got %v", err)
	}
}
