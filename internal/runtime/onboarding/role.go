package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
)

func (w *Workflow) SetRole(ctx context.Context, selection preferences.RoleSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if err := w.preferences.SetRole(ctx, validated); err != nil {
		return w.currentStateWithError(ctx, err)
	}
	return w.State(ctx)
}
