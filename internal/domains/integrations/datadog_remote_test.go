package integrations

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domains/tenancy"
	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/infrastructure/controlplane/api/apitest"
)

func TestRemoteDatadogService_MappingAndValidation(t *testing.T) {
	var nilSvc *RemoteDatadogService
	if _, err := nilSvc.GetByAccount(context.Background(), "acc_1"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected uninitialized error, got %v", err)
	}

	svc := NewRemoteDatadogService(&apitest.Client{})
	if _, err := svc.GetByAccount(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}
	if _, _, err := svc.ValidateAPIKey(context.Background(), "", "key"); err == nil || !strings.Contains(err.Error(), "datadog site is required") {
		t.Fatalf("expected site validation error, got %v", err)
	}
	if _, _, err := svc.ValidateAPIKey(context.Background(), DatadogSiteUS1, ""); err == nil || !strings.Contains(err.Error(), "api key is required") {
		t.Fatalf("expected api key validation error, got %v", err)
	}
	if _, err := svc.Create(context.Background(), CreateDatadogAccountInput{}); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}
	if _, err := svc.Status(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "datadog account id is required") {
		t.Fatalf("expected datadog account id validation error, got %v", err)
	}

	mock := &apitest.Client{
		GetAccountDatadogAccountFn: func(context.Context, controlplane.AccountID) (*controlplane.DatadogAccount, error) {
			return &controlplane.DatadogAccount{ID: "dd_1", Name: "Main", Site: controlplane.DatadogSiteUS1}, nil
		},
		ValidateDatadogAPIKeyFn: func(context.Context, string, controlplane.DatadogSite) (bool, string, error) {
			return true, "", nil
		},
		CreateDatadogAccountWithCredentialsFn: func(_ context.Context, input controlplane.CreateDatadogAccountInput) (controlplane.DatadogAccount, error) {
			if input.AccountID != "acc_1" || input.Name != "Main" || input.APIKey != "api" || input.AppKey != "app" {
				return controlplane.DatadogAccount{}, errors.New("unexpected create input")
			}
			return controlplane.DatadogAccount{ID: "dd_new", Name: input.Name, Site: input.Site}, nil
		},
		GetDatadogAccountStatusFn: func(context.Context, controlplane.DatadogAccountID) (*controlplane.DatadogAccountStatus, error) {
			return &controlplane.DatadogAccountStatus{
				Health:              controlplane.DatadogAccountHealthOK,
				ReadyForUse:         true,
				ServiceCount:        3,
				ActiveServices:      2,
				EventCount:          10,
				AnalyzedCount:       9,
				PendingPolicyCount:  1,
				ApprovedPolicyCount: 2,
			}, nil
		},
	}

	svc = NewRemoteDatadogService(mock)
	account, err := svc.GetByAccount(context.Background(), tenancy.AccountID("acc_1"))
	if err != nil {
		t.Fatalf("get by account: %v", err)
	}
	if account == nil || account.ID != "dd_1" || account.Site != DatadogSiteUS1 {
		t.Fatalf("unexpected account mapping: %+v", account)
	}

	ok, msg, err := svc.ValidateAPIKey(context.Background(), DatadogSiteUS1, "api")
	if err != nil || !ok || msg != "" {
		t.Fatalf("unexpected validate result ok=%v msg=%q err=%v", ok, msg, err)
	}

	createdID, err := svc.Create(context.Background(), CreateDatadogAccountInput{
		AccountID: "acc_1",
		Name:      "Main",
		Site:      DatadogSiteUS1,
		APIKey:    "api",
		AppKey:    "app",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if createdID != "dd_new" {
		t.Fatalf("unexpected created id: %q", createdID)
	}

	status, err := svc.Status(context.Background(), "dd_1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status == nil || !status.ReadyForUse || status.Health != DatadogHealthOK || status.ServiceCount != 3 {
		t.Fatalf("unexpected status mapping: %+v", status)
	}
}
