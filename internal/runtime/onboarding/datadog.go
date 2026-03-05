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

func (s *Service) SubmitDatadogAPIKey(ctx context.Context, apiKey string) (State, error) {
	if apiKey == "" {
		return State{}, fmt.Errorf("api key is required")
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
	valid, message, err := s.datadog.ValidateAPIKey(ctx, state.DatadogDraft.Site, apiKey)
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
		d.apiKey = apiKey
	})
	return s.State(ctx)
}

func (s *Service) SubmitDatadogAppKey(ctx context.Context, name, appKey string) (State, error) {
	if name == "" {
		return State{}, fmt.Errorf("datadog account name is required")
	}
	if appKey == "" {
		return State{}, fmt.Errorf("app key is required")
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

	_, err = s.datadog.Create(ctx, integrations.CreateDatadogAccountInput{
		AccountID: state.SelectedAccount.ID,
		Name:      name,
		Site:      state.DatadogDraft.Site,
		APIKey:    state.DatadogDraft.apiKey,
		AppKey:    appKey,
	})
	if err != nil {
		return State{}, err
	}

	s.setDraft(func(d *DatadogDraft) {
		*d = DatadogDraft{}
	})
	return s.State(ctx)
}
