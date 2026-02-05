package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Organizations provides access to organizations.
type Organizations interface {
	List(ctx context.Context) ([]domain.Organization, error)
	Create(ctx context.Context, id uuid.UUID, name string) (*OrganizationBootstrapResult, error)
}

// OrganizationService handles organization-related API operations.
type OrganizationService struct {
	client Client
	scope  log.Scope
}

// Ensure OrganizationService implements Organizations.
var _ Organizations = (*OrganizationService)(nil)

// NewOrganizationService creates a new organization service.
func NewOrganizationService(client Client, scope log.Scope) *OrganizationService {
	return &OrganizationService{
		client: client,
		scope:  scope.Child("organizations"),
	}
}

// OrganizationBootstrapResult contains the organization, account, and workspace created during bootstrap.
type OrganizationBootstrapResult struct {
	Organization *domain.Organization
	Account      *domain.Account
	Workspace    *domain.Workspace
}

// List fetches all organizations for the user.
func (s *OrganizationService) List(ctx context.Context) ([]domain.Organization, error) {
	s.scope.Debug("fetching organizations from API")
	resp, err := s.client.ListOrganizations(ctx)
	if err != nil {
		s.scope.Error("failed to fetch organizations", "error", err)
		return nil, err
	}

	// Convert GraphQL response to domain model
	orgs := make([]domain.Organization, len(resp.Organizations.Edges))
	for i, edge := range resp.Organizations.Edges {
		orgs[i] = domain.Organization{
			ID:                   domain.OrganizationID(edge.Node.Id),
			Name:                 edge.Node.Name,
			WorkosOrganizationID: domain.WorkosOrganizationID(edge.Node.WorkosOrganizationID),
		}
	}

	s.scope.Debug("fetched organizations from API", "count", len(orgs))
	return orgs, nil
}

// Create creates a new organization with bootstrapped account and workspace.
func (s *OrganizationService) Create(ctx context.Context, id uuid.UUID, name string) (*OrganizationBootstrapResult, error) {
	s.scope.Debug("creating organization with bootstrap via API", "id", id.String(), "name", name)
	input := gen.CreateOrganizationInput{
		Id:   ptr(id.String()),
		Name: name,
	}

	resp, err := s.client.CreateOrganizationAndBootstrap(ctx, input)
	if err != nil {
		s.scope.Error("failed to create organization", "error", err)
		return nil, err
	}

	org := &domain.Organization{
		ID:                   domain.OrganizationID(resp.CreateOrganizationAndBootstrap.Organization.Id),
		Name:                 resp.CreateOrganizationAndBootstrap.Organization.Name,
		WorkosOrganizationID: domain.WorkosOrganizationID(resp.CreateOrganizationAndBootstrap.Organization.WorkosOrganizationID),
	}

	account := &domain.Account{
		ID:   domain.AccountID(resp.CreateOrganizationAndBootstrap.Account.Id),
		Name: resp.CreateOrganizationAndBootstrap.Account.Name,
	}

	workspace := &domain.Workspace{
		ID:   domain.WorkspaceID(resp.CreateOrganizationAndBootstrap.Workspace.Id),
		Name: resp.CreateOrganizationAndBootstrap.Workspace.Name,
	}

	result := &OrganizationBootstrapResult{
		Organization: org,
		Account:      account,
		Workspace:    workspace,
	}

	s.scope.Debug("created organization via API", "id", org.ID, "name", org.Name, "accountID", account.ID)
	return result, nil
}
