package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
)

func (s *Service) SetRole(ctx context.Context, selection preferences.RoleSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetRole(ctx, validated); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
