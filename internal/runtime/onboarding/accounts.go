package onboarding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

func (s *Service) SelectAccount(ctx context.Context, accountID tenancy.AccountID) (State, error) {
	if accountID == "" {
		return State{}, fmt.Errorf("account id is required")
	}
	if err := s.preferences.SetAccount(ctx, accountID); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}

func (s *Service) CreateAccount(ctx context.Context, name string) (State, error) {
	if name == "" {
		return State{}, fmt.Errorf("account name is required")
	}
	state, err := s.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedOrganization == nil {
		return State{}, fmt.Errorf("organization must be selected before creating an account")
	}

	id, err := s.accounts(state.SelectedOrganization.ID).Create(ctx, name)
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetAccount(ctx, id); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
