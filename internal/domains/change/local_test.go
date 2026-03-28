package change

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite/sqlitetest"
)

func TestLocalService_ListsFindings(t *testing.T) {
	db := sqlitetest.Open(t)
	ctx := context.Background()

	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO findings (id, account_id, service_id, log_event_id, domain, scope_kind, type, problem_version, fingerprint, details, created_at)
		VALUES ('f_1', 'acc_1', 'svc_1', 'evt_1', 'quality', 'log_event', 'wrong_severity', 4, 'fp_1', '{"example":true}', '2026-03-17T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert finding: %v", err)
	}
	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO finding_curations (id, account_id, finding_id, finding_problem_version, disposition, title, body, priority, version, created_at)
		VALUES ('fc_1', 'acc_1', 'f_1', 4, 'SURFACE', 'Checkout severity drift', 'Severity does not match failure impact.', 'high', 1, '2026-03-17T00:05:00Z')
	`); err != nil {
		t.Fatalf("insert curation: %v", err)
	}
	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO finding_statuses_cache (id, account_id, finding_id, service_id, log_event_id, scope_kind, plan_status, isolated_events_per_hour, isolated_total_usd_per_hour, log_event_count, finding_updated_at, plan_updated_at)
		VALUES ('fs_1', 'acc_1', 'f_1', 'svc_1', 'evt_1', 'log_event', 'PROPOSED', 12.5, 4.75, 18, '2026-03-17T00:10:00Z', '2026-03-17T00:11:00Z')
	`); err != nil {
		t.Fatalf("insert finding status: %v", err)
	}
	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO finding_plans (id, account_id, finding_id, title, summary, rationale, open_questions, steps, status, revision, version, created_at)
		VALUES ('fp_1', 'acc_1', 'f_1', 'Lower checkout noise', 'Tune the severity mapping and suppress the noisy path.', 'This restores alert quality and reduces distracting traffic.', '[]', '[]', 'PROPOSED', 1, 1, '2026-03-17T00:12:00Z')
	`); err != nil {
		t.Fatalf("insert finding plan: %v", err)
	}

	svc := NewLocalFindingService(db.Raw())

	all, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("list findings: %v", err)
	}
	if len(all) != 1 || all[0].Curation.Title != "Checkout severity drift" {
		t.Fatalf("unexpected findings: %+v", all)
	}

	byDomain, err := svc.ListByDomain(ctx, "quality")
	if err != nil {
		t.Fatalf("list by domain: %v", err)
	}
	if len(byDomain) != 1 || byDomain[0].Domain != "quality" {
		t.Fatalf("unexpected domain findings: %+v", byDomain)
	}

	byService, err := svc.ListByService(ctx, "svc_1")
	if err != nil {
		t.Fatalf("list by service: %v", err)
	}
	if len(byService) != 1 || byService[0].Plan.Status != "PROPOSED" || byService[0].Plan.Title != "Lower checkout noise" {
		t.Fatalf("unexpected service findings: %+v", byService)
	}
}

func TestLocalService_ValidatesScopedReads(t *testing.T) {
	db := sqlitetest.Open(t)
	svc := NewLocalFindingService(db.Raw())

	if _, err := svc.ListByDomain(context.Background(), ""); err == nil {
		t.Fatal("expected domain validation")
	}
	if _, err := svc.ListByService(context.Background(), ""); err == nil {
		t.Fatal("expected service validation")
	}
}
