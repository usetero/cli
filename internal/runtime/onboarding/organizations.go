package onboarding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

func (s *Service) SelectOrganization(ctx context.Context, organizationID tenancy.OrganizationID) (State, error) {
	if organizationID == "" {
		return State{}, fmt.Errorf("organization id is required")
	}
	if err := s.preferences.SetOrganization(ctx, organizationID); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}

func (s *Service) CreateOrganization(ctx context.Context, name string) (State, error) {
	if name == "" {
		return State{}, fmt.Errorf("organization name is required")
	}
	bootstrap, err := s.orgs.Create(ctx, name)
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetScope(ctx, bootstrap.Organization.ID, bootstrap.Account.ID, bootstrap.Workspace.ID); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
