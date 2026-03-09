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
	var calledListOrg controlplane.OrganizationID
	var calledCreateOrg controlplane.OrganizationID
	var calledDeleteID controlplane.AccountID
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
		DeleteAccountFn: func(_ context.Context, accountID controlplane.AccountID) error {
			calledDeleteID = accountID
			return nil
		},
	}

	svc := NewRemoteAccountService(mock, "org_1")
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

	id, err := svc.Create(context.Background(), AccountCreate{Name: "Ops"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id != "acc_new" {
		t.Fatalf("mapped id = %q", id)
	}
	if calledCreateOrg != controlplane.OrganizationID("org_1") || calledCreateName != "Ops" {
		t.Fatalf("create called with org=%q name=%q", calledCreateOrg, calledCreateName)
	}
	if err := svc.Delete(context.Background(), "acc_1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if calledDeleteID != controlplane.AccountID("acc_1") {
		t.Fatalf("delete called with account=%q", calledDeleteID)
	}
	if err := svc.Delete(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}

	mock.ListAccountsFn = func(context.Context, controlplane.OrganizationID) ([]controlplane.Account, error) {
		return nil, errors.New("boom")
	}
	if _, err := svc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}
