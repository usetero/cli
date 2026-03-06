package tenancy

import (
	"context"
	"testing"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/infrastructure/controlplane/api/apitest"
)

func TestRemoteOrganizationService_MappingAndValidation(t *testing.T) {
	svc := NewRemoteOrganizationService(&apitest.Client{})
	if _, err := svc.Create(context.Background(), OrganizationCreate{}); err == nil || err.Error() != "organization name is required" {
		t.Fatalf("expected organization name validation error, got %v", err)
	}

	calledCreateName := ""
	mock := &apitest.Client{
		ListOrganizationsFn: func(context.Context) ([]controlplane.Organization, error) {
			return []controlplane.Organization{{ID: "org_1", Name: "Org", WorkosOrganizationID: "wo_1"}}, nil
		},
		CreateOrganizationAndBootstrapFn: func(_ context.Context, name string) (controlplane.OrganizationBootstrap, error) {
			calledCreateName = name
			return controlplane.OrganizationBootstrap{
				Organization: controlplane.Organization{ID: "org_1", Name: "Org", WorkosOrganizationID: "wo_1"},
				Account:      controlplane.Account{ID: "acc_1", Name: "Primary"},
				Workspace:    controlplane.Workspace{ID: "ws_1", Name: "Default"},
			}, nil
		},
	}
	svc = NewRemoteOrganizationService(mock)

	orgs, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list orgs: %v", err)
	}
	if len(orgs) != 1 || orgs[0].ID != "org_1" || orgs[0].WorkosOrganizationID != "wo_1" {
		t.Fatalf("unexpected mapped orgs: %+v", orgs)
	}

	bootstrap, err := svc.Create(context.Background(), OrganizationCreate{Name: "Org"})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	if calledCreateName != "Org" {
		t.Fatalf("create called with name=%q", calledCreateName)
	}
	if bootstrap.Organization.ID != "org_1" || bootstrap.Account.ID != "acc_1" || bootstrap.Workspace.ID != "ws_1" {
		t.Fatalf("unexpected bootstrap mapping: %+v", bootstrap)
	}
	if bootstrap.Workspace.AccountID != "acc_1" {
		t.Fatalf("expected workspace account id to map from bootstrap account id")
	}
}
