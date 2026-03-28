package onboarding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

func (w *Workflow) SelectAccount(ctx context.Context, selection preferences.AccountSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if err := w.preferences.SetAccount(ctx, validated); err != nil {
		return w.currentStateWithError(ctx, err)
	}
	return w.State(ctx)
}

func (w *Workflow) CreateAccount(ctx context.Context, create tenancy.AccountCreate) (State, error) {
	validated, err := create.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	state, err := w.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedOrganization == nil {
		return w.currentStateWithError(ctx, fmt.Errorf("organization must be selected before creating an account"))
	}

	id, err := w.accounts(state.SelectedOrganization.ID).Create(ctx, validated)
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if err := w.preferences.SetAccount(ctx, preferences.AccountSelection{AccountID: id}); err != nil {
		return w.currentStateWithError(ctx, err)
	}
	return w.State(ctx)
}
