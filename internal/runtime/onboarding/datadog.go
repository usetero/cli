package onboarding

import (
	"context"
	"errors"
	"fmt"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/tenancy"
)

// DatadogDraft stores in-progress Datadog setup while onboarding is incomplete.
type DatadogDraft struct {
	Site      integrations.DatadogSite
	HasAPIKey bool
	apiKey    integrations.DatadogAPIKey
}

func (w *Workflow) currentDraft(accountID tenancy.AccountID) DatadogDraft {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.bound != "" && accountID != "" && w.bound != accountID {
		w.draft = DatadogDraft{}
	}
	w.bound = accountID
	return w.draft
}

func (w *Workflow) setDraft(update func(d *DatadogDraft)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	update(&w.draft)
}

func (w *Workflow) SetDatadogSite(ctx context.Context, site integrations.DatadogSite) (State, error) {
	if !site.Valid() {
		return w.currentStateWithError(ctx, fmt.Errorf("invalid datadog site: %q", site))
	}
	w.setDraft(func(d *DatadogDraft) {
		d.Site = site
		d.HasAPIKey = false
		d.apiKey = ""
	})
	return w.State(ctx)
}

func (w *Workflow) SubmitDatadogAPIKey(ctx context.Context, submission integrations.DatadogAPIKeySubmission) (State, error) {
	validatedSubmission, err := submission.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	state, err := w.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedAccount == nil {
		return w.currentStateWithError(ctx, fmt.Errorf("account must be selected before datadog setup"))
	}
	if !state.DatadogDraft.Site.Valid() {
		return w.currentStateWithError(ctx, fmt.Errorf("datadog site must be selected first"))
	}
	valid, message, err := w.datadog.ValidateAPIKey(ctx, integrations.DatadogAPIKeyValidation{
		Site:   state.DatadogDraft.Site,
		APIKey: validatedSubmission.APIKey,
	})
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	if !valid {
		if message == "" {
			message = "datadog api key is invalid"
		}
		return w.currentStateWithError(ctx, errors.New(message))
	}

	w.setDraft(func(d *DatadogDraft) {
		d.HasAPIKey = true
		d.apiKey = validatedSubmission.APIKey
	})
	return w.State(ctx)
}

func (w *Workflow) SubmitDatadogAppKey(ctx context.Context, submission integrations.DatadogAppKeySubmission) (State, error) {
	validatedSubmission, err := submission.Validate()
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}
	state, err := w.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedAccount == nil {
		return w.currentStateWithError(ctx, fmt.Errorf("account must be selected before datadog setup"))
	}
	if !state.DatadogDraft.Site.Valid() {
		return w.currentStateWithError(ctx, fmt.Errorf("datadog site must be selected first"))
	}
	if !state.DatadogDraft.HasAPIKey {
		return w.currentStateWithError(ctx, fmt.Errorf("datadog api key must be validated first"))
	}

	_, err = w.datadog.Create(ctx, integrations.DatadogAccountCreate{
		AccountID: state.SelectedAccount.ID,
		Name:      validatedSubmission.Name,
		Site:      state.DatadogDraft.Site,
		APIKey:    state.DatadogDraft.apiKey,
		AppKey:    validatedSubmission.AppKey,
	})
	if err != nil {
		return w.currentStateWithError(ctx, err)
	}

	w.setDraft(func(d *DatadogDraft) {
		*d = DatadogDraft{}
	})
	return w.State(ctx)
}
