package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
)

func (w *Workflow) SelectWorkspace(ctx context.Context, selection preferences.WorkspaceSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if err := w.preferences.SetWorkspace(ctx, validated); err != nil {
		return w.currentStateWithError(ctx, err)
	}
	return w.State(ctx)
}
