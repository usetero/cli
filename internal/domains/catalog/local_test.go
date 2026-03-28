package catalog

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite/sqlitetest"
)

func TestLocalService_ListsCatalogEntities(t *testing.T) {
	db := sqlitetest.Open(t)
	ctx := context.Background()

	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO services (id, account_id, name, enabled, initial_weekly_log_count, created_at)
		VALUES ('svc_1', 'acc_1', 'checkout-api', 1, 4200, '2026-03-17T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert service: %v", err)
	}
	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO service_facts (id, account_id, service_id, fact_group, fact_type, namespace, value, version, created_at)
		VALUES ('sf_1', 'acc_1', 'svc_1', 'catalog_service_facts', 'service_profile', 'semantic', '{"summary":"customer checkout api","service_category":"customer_api","primary_responsibilities":["serve checkout traffic"],"system_roles":["request_handling"]}', 3, '2026-03-17T02:00:00Z')
	`); err != nil {
		t.Fatalf("insert service fact: %v", err)
	}
	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO log_events (id, account_id, service_id, name, description, severity, baseline_avg_bytes, baseline_volume_per_hour, created_at)
		VALUES ('evt_1', 'acc_1', 'svc_1', 'checkout.request', 'checkout request log', 'INFO', 128.5, 42.0, '2026-03-17T03:00:00Z')
	`); err != nil {
		t.Fatalf("insert log event: %v", err)
	}
	if _, err := db.Raw().ExecContext(ctx, `
		INSERT INTO log_event_facts (id, account_id, log_event_id, fact_name, slice_name, slice_version, value, created_at)
		VALUES ('lef_1', 'acc_1', 'evt_1', 'identity_profile', 'identity', 7, '{"action":"request","subject_class":"resource","subject":"checkout","outcome":"success","operation":"post_checkout"}', '2026-03-17T04:00:00Z')
	`); err != nil {
		t.Fatalf("insert log event fact: %v", err)
	}

	svc := NewLocalCatalogService(db.Raw())

	services, err := svc.ListServices(ctx)
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	if len(services) != 1 || services[0].Name != "checkout-api" || !services[0].Enabled {
		t.Fatalf("unexpected services: %+v", services)
	}
	if services[0].Facts.ServiceProfile == nil || services[0].Facts.ServiceProfile.ServiceCategory != "customer_api" {
		t.Fatalf("unexpected service facts: %+v", services[0].Facts)
	}

	events, err := svc.ListLogEventsByService(ctx, "svc_1")
	if err != nil {
		t.Fatalf("list log events: %v", err)
	}
	if len(events) != 1 || events[0].Name != "checkout.request" {
		t.Fatalf("unexpected log events: %+v", events)
	}
	if events[0].Facts.IdentityProfile == nil || events[0].Facts.IdentityProfile.Action != "request" {
		t.Fatalf("unexpected event facts: %+v", events[0].Facts)
	}
}

func TestLocalService_RequiresIDsForScopedReads(t *testing.T) {
	db := sqlitetest.Open(t)
	svc := NewLocalCatalogService(db.Raw())

	if _, err := svc.ListLogEventsByService(context.Background(), ""); err == nil {
		t.Fatal("expected service id validation")
	}
}
