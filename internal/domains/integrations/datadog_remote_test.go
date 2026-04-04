package integrations

import (
	"context"
	"errors"
	"strings"
	"testing"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/infrastructure/controlplane/api/apitest"
)

func TestRemoteDatadogService_MappingAndValidation(t *testing.T) {
	svc := NewRemoteDatadogService(&apitest.Client{})
	if _, _, err := svc.ValidateAPIKey(context.Background(), DatadogAPIKeyValidation{APIKey: DatadogAPIKey("key")}); err == nil || !strings.Contains(err.Error(), "datadog site is required") {
		t.Fatalf("expected site validation error, got %v", err)
	}
	if _, _, err := svc.ValidateAPIKey(context.Background(), DatadogAPIKeyValidation{Site: DatadogSiteUS1}); err == nil || !strings.Contains(err.Error(), "api key is required") {
		t.Fatalf("expected api key validation error, got %v", err)
	}
	if _, err := svc.Create(context.Background(), DatadogAccountCreate{}); err == nil || !strings.Contains(err.Error(), "datadog account name is required") {
		t.Fatalf("expected name validation error, got %v", err)
	}
	if _, err := svc.Status(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "datadog account id is required") {
		t.Fatalf("expected datadog account id validation error, got %v", err)
	}

	mock := &apitest.Client{
		GetDatadogAccountFn: func(context.Context) (*controlplane.DatadogAccount, error) {
			return &controlplane.DatadogAccount{ID: "dd_1", Name: "Main", Site: controlplane.DatadogSiteUS1}, nil
		},
		ValidateDatadogAPIKeyFn: func(context.Context, string, controlplane.DatadogSite) (bool, string, error) {
			return true, "", nil
		},
		CreateDatadogAccountWithCredentialsFn: func(_ context.Context, input controlplane.CreateDatadogAccountInput) (controlplane.DatadogAccount, error) {
			if input.Name != "Main" || input.APIKey != "api" || input.AppKey != "app" {
				return controlplane.DatadogAccount{}, errors.New("unexpected create input")
			}
			return controlplane.DatadogAccount{ID: "dd_new", Name: input.Name, Site: input.Site}, nil
		},
		GetDatadogAccountStatusFn: func(context.Context, controlplane.DatadogAccountID) (*controlplane.DatadogAccountStatus, error) {
			return &controlplane.DatadogAccountStatus{
				Health:                 controlplane.DatadogAccountHealthOK,
				ReadyForUse:            true,
				ServiceCount:           3,
				ActiveServices:         2,
				EventCount:             10,
				AnalyzedCount:          9,
				PreviewLogEventCount:   8,
				EffectiveLogEventCount: 6,
			}, nil
		},
	}

	svc = NewRemoteDatadogService(mock)
	account, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("get datadog account: %v", err)
	}
	if account == nil || account.ID != "dd_1" || account.Site != DatadogSiteUS1 {
		t.Fatalf("unexpected account mapping: %+v", account)
	}

	ok, msg, err := svc.ValidateAPIKey(context.Background(), DatadogAPIKeyValidation{Site: DatadogSiteUS1, APIKey: DatadogAPIKey("api")})
	if err != nil || !ok || msg != "" {
		t.Fatalf("unexpected validate result ok=%v msg=%q err=%v", ok, msg, err)
	}

	createdID, err := svc.Create(context.Background(), DatadogAccountCreate{
		Name:   DatadogAccountName("Main"),
		Site:   DatadogSiteUS1,
		APIKey: DatadogAPIKey("api"),
		AppKey: DatadogAppKey("app"),
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
