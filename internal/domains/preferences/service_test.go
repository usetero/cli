package preferences_test

import (
	"context"
	"testing"

	domainprefs "github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/preferences/preferencestest"
)

func TestService_CascadeResets(t *testing.T) {
	store := &preferencestest.Store{}
	svc := domainprefs.NewService(store)
	ctx := context.Background()

	if err := svc.SetOrganization(ctx, "org_1"); err != nil {
		t.Fatalf("set org: %v", err)
	}
	if err := svc.SetAccount(ctx, "acct_1"); err != nil {
		t.Fatalf("set account: %v", err)
	}
	if err := svc.SetWorkspace(ctx, "ws_1"); err != nil {
		t.Fatalf("set workspace: %v", err)
	}

	if err := svc.SetOrganization(ctx, "org_2"); err != nil {
		t.Fatalf("set org: %v", err)
	}

	got, err := svc.Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if got.Organization != tenancy.OrganizationID("org_2") {
		t.Fatalf("org mismatch: %q", got.Organization)
	}
	if got.Account != "" || got.Workspace != "" {
		t.Fatalf("account/workspace should reset after org change: %+v", got)
	}
}

func TestService_SetRoleRejectsInvalid(t *testing.T) {
	svc := domainprefs.NewService(&preferencestest.Store{})
	if err := svc.SetRole(context.Background(), domainprefs.Role("bad")); err == nil {
		t.Fatal("expected error")
	}
}

func TestService_SetScopePersistsAllSelectionsInSingleWrite(t *testing.T) {
	store := &preferencestest.Store{}
	svc := domainprefs.NewService(store)

	if err := svc.SetScope(context.Background(), "org_1", "acct_1", "ws_1"); err != nil {
		t.Fatalf("set scope: %v", err)
	}

	if store.SaveCalls != 1 {
		t.Fatalf("expected one save call, got %d", store.SaveCalls)
	}

	got, err := svc.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if got.Organization != "org_1" || got.Account != "acct_1" || got.Workspace != "ws_1" {
		t.Fatalf("scope mismatch: %+v", got)
	}
}
