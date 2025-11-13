package api

import (
	"context"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/pkg/client"
)

// OrganizationService handles organization-related API operations.
type OrganizationService struct {
	client Client
	logger log.Logger
}

// NewOrganizationService creates a new organization service.
func NewOrganizationService(client Client, logger log.Logger) *OrganizationService {
	return &OrganizationService{
		client: client,
		logger: logger,
	}
}

// Organization is the domain model for an organization.
type Organization struct {
	ID   string
	Name string
}

// OrganizationBootstrapResult contains the organization, account, and workspace created during bootstrap.
type OrganizationBootstrapResult struct {
	Organization *Organization
	Account      *Account
	Workspace    *Workspace
}

// List fetches all organizations for the user.
func (s *OrganizationService) List(ctx context.Context) ([]Organization, error) {
	s.logger.Debug("fetching organizations from API")
	resp, err := s.client.ListOrganizations(ctx)
	if err != nil {
		s.logger.Error("failed to fetch organizations", "error", err)
		return nil, err
	}

	// Convert GraphQL response to domain model
	orgs := make([]Organization, len(resp.Organizations.Edges))
	for i, edge := range resp.Organizations.Edges {
		orgs[i] = Organization{
			ID:   edge.Node.Id,
			Name: edge.Node.Name,
		}
	}

	s.logger.Debug("fetched organizations from API", "count", len(orgs))
	return orgs, nil
}

// Create creates a new organization with bootstrapped account and workspace.
func (s *OrganizationService) Create(ctx context.Context, name string) (*OrganizationBootstrapResult, error) {
	s.logger.Debug("creating organization with bootstrap via API", "name", name)
	input := client.CreateOrganizationInput{
		Name: name,
	}

	resp, err := s.client.CreateOrganizationAndBootstrap(ctx, input)
	if err != nil {
		s.logger.Error("failed to create organization", "error", err)
		return nil, err
	}

	org := &Organization{
		ID:   resp.CreateOrganizationAndBootstrap.Organization.Id,
		Name: resp.CreateOrganizationAndBootstrap.Organization.Name,
	}

	account := &Account{
		ID:   resp.CreateOrganizationAndBootstrap.Account.Id,
		Name: resp.CreateOrganizationAndBootstrap.Account.Name,
	}

	workspace := &Workspace{
		ID:   resp.CreateOrganizationAndBootstrap.Workspace.Id,
		Name: resp.CreateOrganizationAndBootstrap.Workspace.Name,
	}

	result := &OrganizationBootstrapResult{
		Organization: org,
		Account:      account,
		Workspace:    workspace,
	}

	s.logger.Debug("created organization via API", "id", org.ID, "name", org.Name, "accountID", account.ID)
	return result, nil
}
