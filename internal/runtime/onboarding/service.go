package onboarding

import (
	"context"
	"fmt"
	"sync"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

// Service owns onboarding workflow state transitions.
type Service struct {
	preferences preferences.PreferenceService
	orgs        tenancy.OrganizationService
	accounts    func(organizationID tenancy.OrganizationID) tenancy.AccountService
	workspaces  tenancy.WorkspaceService
	datadog     integrations.DatadogService
	readiness   pssyncer.ReadinessService

	mu    sync.Mutex
	draft DatadogDraft
	bound tenancy.AccountID
}

func NewService(
	preferences preferences.PreferenceService,
	orgs tenancy.OrganizationService,
	accounts func(organizationID tenancy.OrganizationID) tenancy.AccountService,
	workspaces tenancy.WorkspaceService,
	datadog integrations.DatadogService,
	readiness pssyncer.ReadinessService,
) (*Service, error) {
	if preferences == nil {
		return nil, fmt.Errorf("onboarding preferences dependency is required")
	}
	if orgs == nil {
		return nil, fmt.Errorf("onboarding organizations dependency is required")
	}
	if accounts == nil {
		return nil, fmt.Errorf("onboarding accounts dependency is required")
	}
	if workspaces == nil {
		return nil, fmt.Errorf("onboarding workspaces dependency is required")
	}
	if datadog == nil {
		return nil, fmt.Errorf("onboarding datadog dependency is required")
	}
	if readiness == nil {
		return nil, fmt.Errorf("onboarding powersync readiness dependency is required")
	}

	return &Service{
		preferences: preferences,
		orgs:        orgs,
		accounts:    accounts,
		workspaces:  workspaces,
		datadog:     datadog,
		readiness:   readiness,
	}, nil
}

// State returns current onboarding projection and next step.
func (s *Service) State(ctx context.Context) (State, error) {
	if err := s.validate(); err != nil {
		return State{}, err
	}

	pref, err := s.preferences.Snapshot(ctx)
	if err != nil {
		return State{}, err
	}

	state, err := s.loadState(ctx, pref)
	if err != nil {
		return State{}, err
	}
	state.NextStep = nextStep(state)
	if state.Role == "" {
		state.NextStep = StepRoleSelect
	}
	return state, nil
}

// Refresh is an alias for State and is useful for polling loops.
func (s *Service) Refresh(ctx context.Context) (State, error) {
	return s.State(ctx)
}

func (s *Service) validate() error {
	if s == nil {
		return fmt.Errorf("onboarding service is nil")
	}
	if s.preferences == nil {
		return fmt.Errorf("onboarding preferences dependency is required")
	}
	if s.orgs == nil {
		return fmt.Errorf("onboarding organizations dependency is required")
	}
	if s.accounts == nil {
		return fmt.Errorf("onboarding accounts dependency is required")
	}
	if s.workspaces == nil {
		return fmt.Errorf("onboarding workspaces dependency is required")
	}
	if s.datadog == nil {
		return fmt.Errorf("onboarding datadog dependency is required")
	}
	if s.readiness == nil {
		return fmt.Errorf("onboarding powersync readiness dependency is required")
	}
	return nil
}

func (s *Service) currentDraft(accountID tenancy.AccountID) DatadogDraft {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bound != "" && accountID != "" && s.bound != accountID {
		s.draft = DatadogDraft{}
	}
	s.bound = accountID
	return s.draft
}

func (s *Service) setDraft(update func(d *DatadogDraft)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	update(&s.draft)
}
