package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

func (w *Workflow) SelectOrganization(ctx context.Context, selection preferences.OrganizationSelection) (State, error) {
	validated, err := selection.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if err := w.preferences.SetOrganization(ctx, validated); err != nil {
		return w.currentStateWithError(ctx, err)
	}
	return w.State(ctx)
}

func (w *Workflow) CreateOrganization(ctx context.Context, create tenancy.OrganizationCreate) (State, error) {
	validated, err := create.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	bootstrap, err := w.orgs.Create(ctx, validated)
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if err := w.preferences.SetScope(ctx, preferences.ScopeSelection{
		OrganizationID: bootstrap.Organization.ID,
		AccountID:      bootstrap.Account.ID,
		WorkspaceID:    bootstrap.Workspace.ID,
	}); err != nil {
		return w.currentStateWithError(ctx, err)
	}
	return w.State(ctx)
}
