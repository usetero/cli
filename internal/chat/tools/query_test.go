package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestCheckQueryPlan(t *testing.T) {
	t.Parallel()

	db := sqlitetest.OpenBareDB(t)
	ctx := context.Background()

	// Simulate PowerSync pattern: underlying tables store JSON, views expose columns.
	// Expression indexes exist but SQLite can't use them through views for JOINs.
	for _, ddl := range []string{
		`CREATE TABLE ps_data__parent (id TEXT PRIMARY KEY, data TEXT)`,
		`CREATE TABLE ps_data__child (id TEXT PRIMARY KEY, data TEXT)`,
		`CREATE INDEX idx_child_parent_id ON ps_data__child (CAST(json_extract(data, '$.parent_id') AS TEXT))`,
		`CREATE VIEW parent AS SELECT id, CAST(json_extract(data, '$.name') AS TEXT) as name FROM ps_data__parent`,
		`CREATE VIEW child AS SELECT id, CAST(json_extract(data, '$.parent_id') AS TEXT) as parent_id, CAST(json_extract(data, '$.value') AS TEXT) as value FROM ps_data__child`,
	} {
		if _, err := db.Raw().ExecContext(ctx, ddl); err != nil {
			t.Fatalf("DDL: %v", err)
		}
	}

	tool := NewQueryTool(db, logtest.NewScope(t))

	t.Run("simple select passes", func(t *testing.T) {
		t.Parallel()
		err := tool.checkQueryPlan(ctx, "SELECT * FROM parent")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("PK join passes", func(t *testing.T) {
		t.Parallel()
		err := tool.checkQueryPlan(ctx,
			"SELECT c.value FROM child c JOIN parent p ON p.id = c.parent_id")
		if err != nil {
			t.Errorf("expected no error for PK join, got: %v", err)
		}
	})

	t.Run("non-PK left join rejected", func(t *testing.T) {
		t.Parallel()
		// LEFT JOIN on child.parent_id through a view — full table scan.
		err := tool.checkQueryPlan(ctx,
			"SELECT p.name FROM parent p LEFT JOIN child c ON c.parent_id = p.id")
		if err == nil {
			t.Fatal("expected error for SCAN join, got nil")
		}
		if !strings.Contains(err.Error(), "query rejected") {
			t.Errorf("expected 'query rejected' in error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "SCAN") {
			t.Errorf("expected 'SCAN' in error details, got: %v", err)
		}
	})

	t.Run("invalid SQL returns nil", func(t *testing.T) {
		t.Parallel()
		err := tool.checkQueryPlan(ctx, "NOT VALID SQL AT ALL")
		if err != nil {
			t.Errorf("expected nil for invalid SQL, got: %v", err)
		}
	})
}

func TestQueryToolExecute_CapsLargeResults(t *testing.T) {
	t.Parallel()

	db := sqlitetest.OpenBareDB(t)
	ctx := context.Background()

	if _, err := db.Raw().ExecContext(ctx, `CREATE TABLE test_rows (id INTEGER PRIMARY KEY, payload TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	payload := strings.Repeat("x", 2000)
	for i := 0; i < 200; i++ {
		if _, err := db.Raw().ExecContext(ctx, `INSERT INTO test_rows (payload) VALUES (?)`, fmt.Sprintf("%s-%d", payload, i)); err != nil {
			t.Fatalf("insert row %d: %v", i, err)
		}
	}

	tool := NewQueryTool(db, logtest.NewScope(t))
	input, err := json.Marshal(map[string]any{
		"sql":    "SELECT id, payload FROM test_rows ORDER BY id",
		"status": "running",
		"result": "done",
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	result, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(result.Rows) == 0 {
		t.Fatalf("expected at least one row")
	}
	if result.RowsDropped == 0 {
		t.Fatalf("expected rows to be dropped by cap")
	}
}
