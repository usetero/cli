package onboarding

import (
	"context"
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
) *Service {
	if preferences == nil {
		panic("onboarding service requires preferences")
	}
	if orgs == nil {
		panic("onboarding service requires organization service")
	}
	if accounts == nil {
		panic("onboarding service requires account factory")
	}
	if workspaces == nil {
		panic("onboarding service requires workspace service")
	}
	if datadog == nil {
		panic("onboarding service requires datadog service")
	}
	if readiness == nil {
		panic("onboarding service requires powersync readiness")
	}

	return &Service{
		preferences: preferences,
		orgs:        orgs,
		accounts:    accounts,
		workspaces:  workspaces,
		datadog:     datadog,
		readiness:   readiness,
	}
}

// State returns current onboarding projection and next step.
func (s *Service) State(ctx context.Context) (State, error) {
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
