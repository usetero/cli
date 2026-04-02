package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/sqlite"
)

func TestDatadogAccountStatusesGetSummary_UsesCurrentSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "summary.sqlite")
	db, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("ApplySchema() error = %v", err)
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO datadog_account_statuses_cache (
			id,
			account_id,
			ready_for_use,
			health,
			log_service_count,
			log_active_services,
			log_event_count,
			log_event_analyzed_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "status-1", "acc-1", 1, "OK", 3, 2, 20, 12); err != nil {
		t.Fatalf("insert datadog_account_statuses_cache: %v", err)
	}

	summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}

	if !summary.ReadyForUse {
		t.Fatal("ReadyForUse = false, want true")
	}
	if summary.ServiceCount != 3 {
		t.Fatalf("ServiceCount = %d, want 3", summary.ServiceCount)
	}
	if summary.ActiveServices != 2 {
		t.Fatalf("ActiveServices = %d, want 2", summary.ActiveServices)
	}
	if summary.EventCount != 20 {
		t.Fatalf("EventCount = %d, want 20", summary.EventCount)
	}
	if summary.AnalyzedCount != 12 {
		t.Fatalf("AnalyzedCount = %d, want 12", summary.AnalyzedCount)
	}
	if summary.PendingPolicyCount != 0 || summary.ApprovedPolicyCount != 0 || summary.DismissedPolicyCount != 0 {
		t.Fatalf("expected zero legacy policy counts, got pending=%d approved=%d dismissed=%d", summary.PendingPolicyCount, summary.ApprovedPolicyCount, summary.DismissedPolicyCount)
	}
}
