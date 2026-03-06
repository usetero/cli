package onboarding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

func (s *Service) SelectAccount(ctx context.Context, selection preferences.AccountSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetAccount(ctx, validated); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}

func (s *Service) CreateAccount(ctx context.Context, create tenancy.AccountCreate) (State, error) {
	validated, err := create.Validate()
	if err != nil {
		return State{}, err
	}
	state, err := s.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedOrganization == nil {
		return State{}, fmt.Errorf("organization must be selected before creating an account")
	}

	id, err := s.accounts(state.SelectedOrganization.ID).Create(ctx, validated)
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetAccount(ctx, preferences.AccountSelection{AccountID: id}); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
