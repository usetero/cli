package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

func (s *Service) SelectOrganization(ctx context.Context, selection preferences.OrganizationSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetOrganization(ctx, validated); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}

func (s *Service) CreateOrganization(ctx context.Context, create tenancy.OrganizationCreate) (State, error) {
	validated, err := create.Validate()
	if err != nil {
		return State{}, err
	}
	bootstrap, err := s.orgs.Create(ctx, validated)
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetScope(ctx, preferences.ScopeSelection{
		OrganizationID: bootstrap.Organization.ID,
		AccountID:      bootstrap.Account.ID,
		WorkspaceID:    bootstrap.Workspace.ID,
	}); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
