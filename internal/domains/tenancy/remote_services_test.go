package tenancy

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/infrastructure/controlplane/api/apitest"
)

func TestRemoteAccountService_MappingAndValidation(t *testing.T) {
	var nilSvc *RemoteAccountService
	if _, err := nilSvc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected uninitialized error, got %v", err)
	}

	svc := NewRemoteAccountService(&apitest.Client{}, "")
	if _, err := svc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "organization id is required") {
		t.Fatalf("expected organization id validation error, got %v", err)
	}

	var calledListOrg controlplane.OrganizationID
	var calledCreateOrg controlplane.OrganizationID
	calledCreateName := ""
	now := time.Now().UTC()
	mock := &apitest.Client{
		ListAccountsFn: func(_ context.Context, organizationID controlplane.OrganizationID) ([]controlplane.Account, error) {
			calledListOrg = organizationID
			return []controlplane.Account{{ID: "acc_1", Name: "Primary", CreatedAt: now}}, nil
		},
		CreateAccountFn: func(_ context.Context, organizationID controlplane.OrganizationID, name string) (controlplane.Account, error) {
			calledCreateOrg = organizationID
			calledCreateName = name
			return controlplane.Account{ID: "acc_new", Name: name, CreatedAt: now}, nil
		},
	}

	svc = NewRemoteAccountService(mock, "org_1")
	accounts, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if calledListOrg != controlplane.OrganizationID("org_1") {
		t.Fatalf("list called with org=%q", calledListOrg)
	}
	if len(accounts) != 1 || accounts[0].ID != "acc_1" || accounts[0].Name != "Primary" {
		t.Fatalf("unexpected mapped accounts: %+v", accounts)
	}

	id, err := svc.Create(context.Background(), "Ops")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id != "acc_new" {
		t.Fatalf("mapped id = %q", id)
	}
	if calledCreateOrg != controlplane.OrganizationID("org_1") || calledCreateName != "Ops" {
		t.Fatalf("create called with org=%q name=%q", calledCreateOrg, calledCreateName)
	}

	mock.ListAccountsFn = func(context.Context, controlplane.OrganizationID) ([]controlplane.Account, error) {
		return nil, errors.New("boom")
	}
	if _, err := svc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}

func TestRemoteOrganizationService_MappingAndValidation(t *testing.T) {
	var nilSvc *RemoteOrganizationService
	if _, err := nilSvc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected uninitialized error, got %v", err)
	}
	if _, err := nilSvc.Create(context.Background(), "Org"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected uninitialized error, got %v", err)
	}

	svc := NewRemoteOrganizationService(&apitest.Client{})
	if _, err := svc.Create(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "organization name is required") {
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

	bootstrap, err := svc.Create(context.Background(), "Org")
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

func TestRemoteWorkspaceService_MappingAndValidation(t *testing.T) {
	var nilSvc *RemoteWorkspaceService
	if _, err := nilSvc.ListByAccount(context.Background(), "acc_1"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected uninitialized error, got %v", err)
	}

	svc := NewRemoteWorkspaceService(&apitest.Client{})
	if _, err := svc.ListByAccount(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}

	var calledAccountID controlplane.AccountID
	now := time.Now().UTC()
	mock := &apitest.Client{
		ListWorkspacesFn: func(_ context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error) {
			calledAccountID = accountID
			return []controlplane.Workspace{{ID: "ws_1", Name: "Default", CreatedAt: now}}, nil
		},
	}
	svc = NewRemoteWorkspaceService(mock)
	rows, err := svc.ListByAccount(context.Background(), "acc_1")
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if calledAccountID != controlplane.AccountID("acc_1") {
		t.Fatalf("list called with account=%q", calledAccountID)
	}
	if len(rows) != 1 || rows[0].ID != "ws_1" || rows[0].AccountID != "acc_1" {
		t.Fatalf("unexpected workspace mapping: %+v", rows)
	}

	if _, err := svc.Create(context.Background(), "acc_1", "Default"); err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("expected not implemented create error, got %v", err)
	}
	if err := svc.Delete(context.Background(), "ws_1"); err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("expected not implemented delete error, got %v", err)
	}
}
