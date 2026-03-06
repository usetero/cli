package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
)

func (s *Service) SelectWorkspace(ctx context.Context, selection preferences.WorkspaceSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return State{}, err
	}
	if err := s.preferences.SetWorkspace(ctx, validated); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
