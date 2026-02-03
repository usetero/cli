package query_test

import (
	"encoding/json"
	"testing"

	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/tui/app/tools/query"
)

func TestTool_Execute(t *testing.T) {
	t.Parallel()

	t.Run("returns query results", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		ctx := t.Context()

		// Create a test table and insert data
		_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("CREATE TABLE: %v", err)
		}
		_, err = db.Exec(ctx, "INSERT INTO items (id, name) VALUES (1, 'foo'), (2, 'bar')")
		if err != nil {
			t.Fatalf("INSERT: %v", err)
		}

		tool := query.Tool{DB: db}
		input, _ := json.Marshal(map[string]string{"sql": "SELECT id, name FROM items ORDER BY id"})

		result, err := tool.Execute(input)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		rows, ok := result.([]map[string]any)
		if !ok {
			t.Fatalf("result type = %T, want []map[string]any", result)
		}

		if len(rows) != 2 {
			t.Fatalf("len(rows) = %d, want 2", len(rows))
		}

		if rows[0]["name"] != "foo" {
			t.Errorf("rows[0][name] = %v, want foo", rows[0]["name"])
		}
		if rows[1]["name"] != "bar" {
			t.Errorf("rows[1][name] = %v, want bar", rows[1]["name"])
		}
	})

	t.Run("returns empty slice for no results", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		ctx := t.Context()

		_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY)")
		if err != nil {
			t.Fatalf("CREATE TABLE: %v", err)
		}

		tool := query.Tool{DB: db}
		input, _ := json.Marshal(map[string]string{"sql": "SELECT * FROM items"})

		result, err := tool.Execute(input)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		rows, ok := result.([]map[string]any)
		if !ok {
			t.Fatalf("result type = %T, want []map[string]any", result)
		}

		if rows != nil && len(rows) != 0 {
			t.Errorf("len(rows) = %d, want 0", len(rows))
		}
	})

	t.Run("rejects write operations", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		ctx := t.Context()

		_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("CREATE TABLE: %v", err)
		}

		tool := query.Tool{DB: db}
		input, _ := json.Marshal(map[string]string{"sql": "INSERT INTO items (name) VALUES ('hacked')"})

		_, err = tool.Execute(input)
		if err == nil {
			t.Fatal("Execute() expected error for INSERT, got nil")
		}
	})

	t.Run("returns error for invalid SQL", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		tool := query.Tool{DB: db}
		input, _ := json.Marshal(map[string]string{"sql": "SELECT * FROM nonexistent"})

		_, err := tool.Execute(input)
		if err == nil {
			t.Fatal("Execute() expected error for nonexistent table, got nil")
		}
	})

	t.Run("returns error for invalid JSON input", func(t *testing.T) {
		t.Parallel()

		db := dbtest.OpenTestDB(t)
		tool := query.Tool{DB: db}

		_, err := tool.Execute([]byte("not json"))
		if err == nil {
			t.Fatal("Execute() expected error for invalid JSON, got nil")
		}
	})
}

func TestTool_Definition(t *testing.T) {
	t.Parallel()

	tool := query.Tool{}
	def := tool.Definition()

	if def.Name != "query" {
		t.Errorf("Name = %q, want query", def.Name)
	}

	if def.InputSchema.Type != "object" {
		t.Errorf("InputSchema.Type = %q, want object", def.InputSchema.Type)
	}

	sqlProp, ok := def.InputSchema.Properties["sql"]
	if !ok {
		t.Fatal("InputSchema.Properties missing 'sql'")
	}
	if sqlProp.Type != "string" {
		t.Errorf("sql.Type = %q, want string", sqlProp.Type)
	}
}
