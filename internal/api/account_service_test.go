package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/pkg/client"
)

func TestAccountService_List(t *testing.T) {
	t.Run("transforms GraphQL edges to domain accounts", func(t *testing.T) {
		mockClient := &apitest.MockClient{
			ListAccountsFunc: func(ctx context.Context, orgID string) (*client.ListAccountsResponse, error) {
				return &client.ListAccountsResponse{
					Accounts: client.ListAccountsAccountsAccountConnection{
						Edges: []client.ListAccountsAccountsAccountConnectionEdgesAccountEdge{
							{Node: client.ListAccountsAccountsAccountConnectionEdgesAccountEdgeNodeAccount{Id: "acc-1", Name: "Production"}},
							{Node: client.ListAccountsAccountsAccountConnectionEdgesAccountEdgeNodeAccount{Id: "acc-2", Name: "Staging"}},
						},
					},
				}, nil
			},
		}

		svc := api.NewAccountService(mockClient, logtest.New(t))
		accounts, err := svc.List(context.Background(), "org-123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(accounts) != 2 {
			t.Fatalf("expected 2 accounts, got %d", len(accounts))
		}
		if accounts[0].ID != "acc-1" || accounts[0].Name != "Production" {
			t.Errorf("first account = %+v, want ID=acc-1, Name=Production", accounts[0])
		}
		if accounts[1].ID != "acc-2" || accounts[1].Name != "Staging" {
			t.Errorf("second account = %+v, want ID=acc-2, Name=Staging", accounts[1])
		}
	})

	t.Run("returns empty slice when no accounts", func(t *testing.T) {
		mockClient := &apitest.MockClient{
			ListAccountsFunc: func(ctx context.Context, orgID string) (*client.ListAccountsResponse, error) {
				return &client.ListAccountsResponse{
					Accounts: client.ListAccountsAccountsAccountConnection{
						Edges: []client.ListAccountsAccountsAccountConnectionEdgesAccountEdge{},
					},
				}, nil
			},
		}

		svc := api.NewAccountService(mockClient, logtest.New(t))
		accounts, err := svc.List(context.Background(), "org-123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(accounts) != 0 {
			t.Errorf("expected 0 accounts, got %d", len(accounts))
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		mockClient := &apitest.MockClient{
			ListAccountsFunc: func(ctx context.Context, orgID string) (*client.ListAccountsResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := api.NewAccountService(mockClient, logtest.New(t))
		_, err := svc.List(context.Background(), "org-123")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "network error" {
			t.Errorf("error = %q, want %q", err.Error(), "network error")
		}
	})
}

func TestAccountService_Create(t *testing.T) {
	t.Run("creates account and returns domain model", func(t *testing.T) {
		var capturedInput client.CreateAccountInput
		mockClient := &apitest.MockClient{
			CreateAccountFunc: func(ctx context.Context, input client.CreateAccountInput) (*client.CreateAccountResponse, error) {
				capturedInput = input
				return &client.CreateAccountResponse{
					CreateAccount: client.CreateAccountCreateAccount{
						Id:   "acc-new",
						Name: "New Account",
					},
				}, nil
			},
		}

		svc := api.NewAccountService(mockClient, logtest.New(t))
		account, err := svc.Create(context.Background(), "org-123", "New Account")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if account.ID != "acc-new" || account.Name != "New Account" {
			t.Errorf("account = %+v, want ID=acc-new, Name=New Account", account)
		}
		if capturedInput.OrganizationID != "org-123" {
			t.Errorf("input.OrganizationID = %q, want %q", capturedInput.OrganizationID, "org-123")
		}
		if capturedInput.Name != "New Account" {
			t.Errorf("input.Name = %q, want %q", capturedInput.Name, "New Account")
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		mockClient := &apitest.MockClient{
			CreateAccountFunc: func(ctx context.Context, input client.CreateAccountInput) (*client.CreateAccountResponse, error) {
				return nil, errors.New("validation error")
			},
		}

		svc := api.NewAccountService(mockClient, logtest.New(t))
		_, err := svc.Create(context.Background(), "org-123", "Test")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
