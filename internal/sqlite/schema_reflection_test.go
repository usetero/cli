package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/sqlite"
)

func TestReflectedSchemaIncludesCurrentSyncedTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "schema.sqlite")
	db, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("ApplySchema() error = %v", err)
	}

	requiredTables := []string{
		"log_events",
		"findings",
		"finding_curations",
		"finding_plans",
		"finding_log_events",
		"finding_statuses_cache",
		"teams",
		"team_memberships",
		"service_team_mappings",
	}

	for _, table := range requiredTables {
		var count int64
		query := "SELECT COUNT(*) FROM pragma_table_info('" + table + "')"
		if err := db.QueryRow(ctx, query).Scan(&count); err != nil {
			t.Fatalf("pragma_table_info(%s): %v", table, err)
		}
		if count == 0 {
			t.Fatalf("expected reflected schema for %s", table)
		}
	}

	var workspaceColumns int64
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM pragma_table_info('conversations') WHERE name = 'workspace_id'").Scan(&workspaceColumns); err != nil {
		t.Fatalf("count conversation workspace column: %v", err)
	}
	if workspaceColumns != 0 {
		t.Fatalf("expected conversations.workspace_id to be absent, found %d columns", workspaceColumns)
	}
}
