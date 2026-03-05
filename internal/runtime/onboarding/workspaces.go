package onboarding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

func (s *Service) SelectWorkspace(ctx context.Context, workspaceID tenancy.WorkspaceID) (State, error) {
	if workspaceID == "" {
		return State{}, fmt.Errorf("workspace id is required")
	}
	if err := s.preferences.SetWorkspace(ctx, workspaceID); err != nil {
		return State{}, err
	}
	return s.State(ctx)
}
