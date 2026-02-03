package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestOrganizationService_List(t *testing.T) {
	t.Parallel()
	t.Run("transforms GraphQL edges to domain organizations with WorkosOrganizationID", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ListOrganizationsFunc: func(ctx context.Context) (*gen.ListOrganizationsResponse, error) {
				return &gen.ListOrganizationsResponse{
					Organizations: gen.ListOrganizationsOrganizationsOrganizationConnection{
						Edges: []*gen.ListOrganizationsOrganizationsOrganizationConnectionEdgesOrganizationEdge{
							{Node: &gen.ListOrganizationsOrganizationsOrganizationConnectionEdgesOrganizationEdgeNodeOrganization{
								Id:                   "org-1",
								Name:                 "Acme Corp",
								WorkosOrganizationID: "org_workos_123",
							}},
							{Node: &gen.ListOrganizationsOrganizationsOrganizationConnectionEdgesOrganizationEdgeNodeOrganization{
								Id:                   "org-2",
								Name:                 "Beta Inc",
								WorkosOrganizationID: "org_workos_456",
							}},
						},
					},
				}, nil
			},
		}

		svc := api.NewOrganizationService(mockClient, logtest.New(t))
		orgs, err := svc.List(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(orgs) != 2 {
			t.Fatalf("expected 2 organizations, got %d", len(orgs))
		}
		if orgs[0].ID != "org-1" || orgs[0].Name != "Acme Corp" || orgs[0].WorkosOrganizationID != "org_workos_123" {
			t.Errorf("first org = %+v, want ID=org-1, Name=Acme Corp, WorkosOrganizationID=org_workos_123", orgs[0])
		}
		if orgs[1].ID != "org-2" || orgs[1].Name != "Beta Inc" || orgs[1].WorkosOrganizationID != "org_workos_456" {
			t.Errorf("second org = %+v, want ID=org-2, Name=Beta Inc, WorkosOrganizationID=org_workos_456", orgs[1])
		}
	})

	t.Run("returns empty slice when no organizations", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ListOrganizationsFunc: func(ctx context.Context) (*gen.ListOrganizationsResponse, error) {
				return &gen.ListOrganizationsResponse{
					Organizations: gen.ListOrganizationsOrganizationsOrganizationConnection{
						Edges: []*gen.ListOrganizationsOrganizationsOrganizationConnectionEdgesOrganizationEdge{},
					},
				}, nil
			},
		}

		svc := api.NewOrganizationService(mockClient, logtest.New(t))
		orgs, err := svc.List(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(orgs) != 0 {
			t.Errorf("expected 0 organizations, got %d", len(orgs))
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ListOrganizationsFunc: func(ctx context.Context) (*gen.ListOrganizationsResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := api.NewOrganizationService(mockClient, logtest.New(t))
		_, err := svc.List(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "network error" {
			t.Errorf("error = %q, want %q", err.Error(), "network error")
		}
	})
}

func TestOrganizationService_Create(t *testing.T) {
	t.Parallel()
	t.Run("creates organization and returns bootstrap result", func(t *testing.T) {
		t.Parallel()
		var capturedInput gen.CreateOrganizationInput
		mockClient := &apitest.MockClient{
			CreateOrganizationAndBootstrapFunc: func(ctx context.Context, input gen.CreateOrganizationInput) (*gen.CreateOrganizationAndBootstrapResponse, error) {
				capturedInput = input
				return &gen.CreateOrganizationAndBootstrapResponse{
					CreateOrganizationAndBootstrap: gen.CreateOrganizationAndBootstrapCreateOrganizationAndBootstrapOrganizationBootstrapResult{
						Organization: gen.CreateOrganizationAndBootstrapCreateOrganizationAndBootstrapOrganizationBootstrapResultOrganization{
							Id:                   "org-new",
							Name:                 "New Org",
							WorkosOrganizationID: "org_workos_new",
						},
						Account: gen.CreateOrganizationAndBootstrapCreateOrganizationAndBootstrapOrganizationBootstrapResultAccount{
							Id:   "acc-new",
							Name: "Default Account",
						},
						Workspace: gen.CreateOrganizationAndBootstrapCreateOrganizationAndBootstrapOrganizationBootstrapResultWorkspace{
							Id:   "ws-new",
							Name: "Default Workspace",
						},
					},
				}, nil
			},
		}

		svc := api.NewOrganizationService(mockClient, logtest.New(t))
		testID := uuid.New()
		result, err := svc.Create(context.Background(), testID, "New Org")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Organization.ID != "org-new" || result.Organization.Name != "New Org" || result.Organization.WorkosOrganizationID != "org_workos_new" {
			t.Errorf("organization = %+v, want ID=org-new, Name=New Org, WorkosOrganizationID=org_workos_new", result.Organization)
		}
		if result.Account.ID != "acc-new" || result.Account.Name != "Default Account" {
			t.Errorf("account = %+v, want ID=acc-new, Name=Default Account", result.Account)
		}
		if result.Workspace.ID != "ws-new" || result.Workspace.Name != "Default Workspace" {
			t.Errorf("workspace = %+v, want ID=ws-new, Name=Default Workspace", result.Workspace)
		}
		if capturedInput.Id == nil || *capturedInput.Id != testID.String() {
			t.Errorf("input.Id = %v, want %q", capturedInput.Id, testID.String())
		}
		if capturedInput.Name != "New Org" {
			t.Errorf("input.Name = %q, want %q", capturedInput.Name, "New Org")
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			CreateOrganizationAndBootstrapFunc: func(ctx context.Context, input gen.CreateOrganizationInput) (*gen.CreateOrganizationAndBootstrapResponse, error) {
				return nil, errors.New("validation error")
			},
		}

		svc := api.NewOrganizationService(mockClient, logtest.New(t))
		_, err := svc.Create(context.Background(), uuid.New(), "Test")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
