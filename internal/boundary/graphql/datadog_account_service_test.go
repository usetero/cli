package graphql_test

import (
	"context"
	"errors"
	"testing"

	api "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestDatadogAccountService_HasAccount(t *testing.T) {
	t.Parallel()
	t.Run("returns true when datadog account exists", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetAccountFunc: func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
				return &gen.GetAccountResponse{
					Accounts: gen.GetAccountAccountsAccountConnection{
						Edges: []*gen.GetAccountAccountsAccountConnectionEdgesAccountEdge{
							{
								Node: &gen.GetAccountAccountsAccountConnectionEdgesAccountEdgeNodeAccount{
									Id: "acc-1",
									DatadogAccount: &gen.GetAccountAccountsAccountConnectionEdgesAccountEdgeNodeAccountDatadogAccount{
										Id:   "dd-123",
										Name: "Production DD",
										Site: gen.DatadogAccountSiteUs1,
									},
								},
							},
						},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		hasAccount, err := svc.HasAccount(context.Background(), domain.AccountID("acc-1"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasAccount {
			t.Error("expected HasAccount to return true")
		}
	})

	t.Run("returns false when datadog account is empty", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetAccountFunc: func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
				return &gen.GetAccountResponse{
					Accounts: gen.GetAccountAccountsAccountConnection{
						Edges: []*gen.GetAccountAccountsAccountConnectionEdgesAccountEdge{
							{
								Node: &gen.GetAccountAccountsAccountConnectionEdgesAccountEdgeNodeAccount{
									Id: "acc-1",
									// Empty DatadogAccount - nil pointer
									DatadogAccount: nil,
								},
							},
						},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		hasAccount, err := svc.HasAccount(context.Background(), domain.AccountID("acc-1"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hasAccount {
			t.Error("expected HasAccount to return false")
		}
	})

	t.Run("returns false when no account found", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetAccountFunc: func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
				return &gen.GetAccountResponse{
					Accounts: gen.GetAccountAccountsAccountConnection{
						Edges: []*gen.GetAccountAccountsAccountConnectionEdgesAccountEdge{},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		hasAccount, err := svc.HasAccount(context.Background(), domain.AccountID("acc-1"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hasAccount {
			t.Error("expected HasAccount to return false")
		}
	})
}

func TestDatadogAccountService_GetAccount(t *testing.T) {
	t.Parallel()
	t.Run("returns datadog account when exists", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetAccountFunc: func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
				return &gen.GetAccountResponse{
					Accounts: gen.GetAccountAccountsAccountConnection{
						Edges: []*gen.GetAccountAccountsAccountConnectionEdgesAccountEdge{
							{
								Node: &gen.GetAccountAccountsAccountConnectionEdgesAccountEdgeNodeAccount{
									Id: "acc-1",
									DatadogAccount: &gen.GetAccountAccountsAccountConnectionEdgesAccountEdgeNodeAccountDatadogAccount{
										Id:   "dd-123",
										Name: "Production DD",
										Site: gen.DatadogAccountSiteUs1,
									},
								},
							},
						},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		account, err := svc.GetAccount(context.Background(), domain.AccountID("acc-1"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if account == nil {
			t.Fatal("expected account, got nil")
		}
		if account.ID != domain.DatadogAccountID("dd-123") {
			t.Errorf("account.ID = %q, want %q", account.ID, "dd-123")
		}
		if account.Name != "Production DD" {
			t.Errorf("account.Name = %q, want %q", account.Name, "Production DD")
		}
		if account.Site != "US1" {
			t.Errorf("account.Site = %q, want %q", account.Site, "US1")
		}
	})

	t.Run("returns nil when no datadog account", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetAccountFunc: func(ctx context.Context, accountID string) (*gen.GetAccountResponse, error) {
				return &gen.GetAccountResponse{
					Accounts: gen.GetAccountAccountsAccountConnection{
						Edges: []*gen.GetAccountAccountsAccountConnectionEdgesAccountEdge{
							{
								Node: &gen.GetAccountAccountsAccountConnectionEdgesAccountEdgeNodeAccount{
									Id:             "acc-1",
									DatadogAccount: nil,
								},
							},
						},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		account, err := svc.GetAccount(context.Background(), domain.AccountID("acc-1"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if account != nil {
			t.Errorf("expected nil, got %+v", account)
		}
	})
}

func TestDatadogAccountService_ValidateAPIKey(t *testing.T) {
	t.Parallel()
	t.Run("returns true for valid key", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ValidateDatadogApiKeyFunc: func(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error) {
				return &gen.ValidateDatadogApiKeyResponse{
					ValidateDatadogApiKey: gen.ValidateDatadogApiKeyValidateDatadogApiKeyValidateDatadogApiKeyResult{
						Valid: true,
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		valid, errMsg, err := svc.ValidateAPIKey(context.Background(), api.ValidateAPIKeyInput{APIKey: "api-key", Site: "US1"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !valid {
			t.Error("expected valid=true")
		}
		if errMsg != "" {
			t.Errorf("expected empty error message, got %q", errMsg)
		}
	})

	t.Run("returns false with error message for invalid key", func(t *testing.T) {
		t.Parallel()
		errMsg := "API key not found"
		mockClient := &apitest.MockClient{
			ValidateDatadogApiKeyFunc: func(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error) {
				return &gen.ValidateDatadogApiKeyResponse{
					ValidateDatadogApiKey: gen.ValidateDatadogApiKeyValidateDatadogApiKeyValidateDatadogApiKeyResult{
						Valid: false,
						Error: &errMsg,
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		valid, errMsg, err := svc.ValidateAPIKey(context.Background(), api.ValidateAPIKeyInput{APIKey: "bad-key", Site: "US1"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if valid {
			t.Error("expected valid=false")
		}
		if errMsg != "API key not found" {
			t.Errorf("errMsg = %q, want %q", errMsg, "API key not found")
		}
	})

	t.Run("returns default error message when none provided", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ValidateDatadogApiKeyFunc: func(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error) {
				return &gen.ValidateDatadogApiKeyResponse{
					ValidateDatadogApiKey: gen.ValidateDatadogApiKeyValidateDatadogApiKeyValidateDatadogApiKeyResult{
						Valid: false,
						Error: nil,
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		valid, errMsg, err := svc.ValidateAPIKey(context.Background(), api.ValidateAPIKeyInput{APIKey: "bad-key", Site: "US1"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if valid {
			t.Error("expected valid=false")
		}
		if errMsg != "Invalid API key" {
			t.Errorf("errMsg = %q, want %q", errMsg, "Invalid API key")
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ValidateDatadogApiKeyFunc: func(ctx context.Context, input gen.ValidateDatadogApiKeyInput) (*gen.ValidateDatadogApiKeyResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		_, _, err := svc.ValidateAPIKey(context.Background(), api.ValidateAPIKeyInput{APIKey: "key", Site: "US1"})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDatadogAccountService_GetStatus(t *testing.T) {
	t.Parallel()
	t.Run("returns status with health and counts", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetDatadogAccountStatusFunc: func(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error) {
				return &gen.GetDatadogAccountStatusResponse{
					DatadogAccounts: gen.GetDatadogAccountStatusDatadogAccountsDatadogAccountConnection{
						Edges: []*gen.GetDatadogAccountStatusDatadogAccountsDatadogAccountConnectionEdgesDatadogAccountEdge{
							{
								Node: &gen.GetDatadogAccountStatusDatadogAccountsDatadogAccountConnectionEdgesDatadogAccountEdgeNodeDatadogAccount{
									Id: "dd-123",
									Status: &gen.GetDatadogAccountStatusDatadogAccountsDatadogAccountConnectionEdgesDatadogAccountEdgeNodeDatadogAccountStatusDatadogAccountStatusCache{
										Health:                gen.DatadogAccountStatusCacheHealthOk,
										ReadyForUse:           true,
										LogServiceCount:       10,
										LogActiveServices:     8,
										OkServices:            7,
										DisabledServices:      1,
										InactiveServices:      1,
										LogEventCount:         200,
										LogEventAnalyzedCount: 180,
										PolicyPendingCount:    12,
									},
								},
							},
						},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		status, err := svc.GetStatus(context.Background(), domain.DatadogAccountID("dd-123"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status == nil {
			t.Fatal("expected status, got nil")
		}
		if status.Health != api.DatadogAccountHealthOK {
			t.Errorf("Health = %q, want %q", status.Health, api.DatadogAccountHealthOK)
		}
		if status.ServiceCount != 10 {
			t.Errorf("ServiceCount = %d, want 10", status.ServiceCount)
		}
		if status.OkServices != 7 {
			t.Errorf("OkServices = %d, want 7", status.OkServices)
		}
		if status.EventCount != 200 {
			t.Errorf("EventCount = %d, want 200", status.EventCount)
		}
		if status.AnalyzedCount != 180 {
			t.Errorf("AnalyzedCount = %d, want 180", status.AnalyzedCount)
		}
		if status.PendingPolicyCount != 12 {
			t.Errorf("PendingPolicyCount = %d, want 12", status.PendingPolicyCount)
		}
	})

	t.Run("returns nil when no datadog account found", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetDatadogAccountStatusFunc: func(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error) {
				return &gen.GetDatadogAccountStatusResponse{
					DatadogAccounts: gen.GetDatadogAccountStatusDatadogAccountsDatadogAccountConnection{
						Edges: []*gen.GetDatadogAccountStatusDatadogAccountsDatadogAccountConnectionEdgesDatadogAccountEdge{},
					},
				}, nil
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		status, err := svc.GetStatus(context.Background(), domain.DatadogAccountID("dd-123"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != nil {
			t.Errorf("expected nil, got %+v", status)
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetDatadogAccountStatusFunc: func(ctx context.Context, id string) (*gen.GetDatadogAccountStatusResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := api.NewDatadogAccountService(mockClient, logtest.NewScope(t))
		_, err := svc.GetStatus(context.Background(), domain.DatadogAccountID("dd-123"))

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
