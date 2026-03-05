package onboarding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/preferences"
)

func (s *Service) SetRole(ctx context.Context, role preferences.Role) (State, error) {
	if !role.Valid() {
		return State{}, fmt.Errorf("invalid role: %q", role)
	}
	if err := s.preferences.SetRole(ctx, role); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
