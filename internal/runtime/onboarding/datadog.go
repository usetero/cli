package onboarding

import (
	"context"
	"errors"
	"fmt"

	"github.com/usetero/cli/internal/domains/integrations"
)

func (s *Service) SetDatadogSite(ctx context.Context, site integrations.DatadogSite) (State, error) {
	if !site.Valid() {
		return State{}, fmt.Errorf("invalid datadog site: %q", site)
	}
	s.setDraft(func(d *DatadogDraft) {
		d.Site = site
		d.HasAPIKey = false
		d.apiKey = ""
	})
	return s.State(ctx)
}

func (s *Service) SubmitDatadogAPIKey(ctx context.Context, submission integrations.DatadogAPIKeySubmission) (State, error) {
	validatedSubmission, err := submission.Validate()
	if err != nil {
		return State{}, err
	}
	state, err := s.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedAccount == nil {
		return State{}, fmt.Errorf("account must be selected before datadog setup")
	}
	if !state.DatadogDraft.Site.Valid() {
		return State{}, fmt.Errorf("datadog site must be selected first")
	}
	valid, message, err := s.datadog.ValidateAPIKey(ctx, integrations.DatadogAPIKeyValidation{
		Site:   state.DatadogDraft.Site,
		APIKey: validatedSubmission.APIKey,
	})
	if err != nil {
		return State{}, err
	}
	if !valid {
		if message == "" {
			message = "datadog api key is invalid"
		}
		return State{}, errors.New(message)
	}

	s.setDraft(func(d *DatadogDraft) {
		d.HasAPIKey = true
		d.apiKey = validatedSubmission.APIKey
	})
	return s.State(ctx)
}

func (s *Service) SubmitDatadogAppKey(ctx context.Context, submission integrations.DatadogAppKeySubmission) (State, error) {
	validatedSubmission, err := submission.Validate()
	if err != nil {
		return State{}, err
	}
	state, err := s.State(ctx)
	if err != nil {
		return State{}, err
	}
	if state.SelectedAccount == nil {
		return State{}, fmt.Errorf("account must be selected before datadog setup")
	}
	if !state.DatadogDraft.Site.Valid() {
		return State{}, fmt.Errorf("datadog site must be selected first")
	}
	if !state.DatadogDraft.HasAPIKey {
		return State{}, fmt.Errorf("datadog api key must be validated first")
	}

	_, err = s.datadog.Create(ctx, integrations.DatadogAccountCreate{
		AccountID: state.SelectedAccount.ID,
		Name:      validatedSubmission.Name,
		Site:      state.DatadogDraft.Site,
		APIKey:    state.DatadogDraft.apiKey,
		AppKey:    validatedSubmission.AppKey,
	})
	if err != nil {
		return State{}, err
	}

	s.setDraft(func(d *DatadogDraft) {
		*d = DatadogDraft{}
	})
	return s.State(ctx)
}
