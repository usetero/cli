package onboarding

import (
	"context"
	"sync"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

type readinessReader interface {
	Ready(ctx context.Context) (bool, error)
}

// Workflow owns onboarding state projection and progression.
type Workflow struct {
	preferences preferences.PreferenceService
	orgs        tenancy.OrganizationService
	accounts    func(organizationID tenancy.OrganizationID) tenancy.AccountService
	workspaces  tenancy.WorkspaceService
	datadog     integrations.DatadogService
	readiness   readinessReader

	mu    sync.Mutex
	draft DatadogDraft
	bound tenancy.AccountID
}

func NewWorkflow(
	preferences preferences.PreferenceService,
	orgs tenancy.OrganizationService,
	accounts func(organizationID tenancy.OrganizationID) tenancy.AccountService,
	workspaces tenancy.WorkspaceService,
	datadog integrations.DatadogService,
	readiness readinessReader,
) *Workflow {
	if preferences == nil {
		panic("onboarding workflow requires preferences")
	}
	if orgs == nil {
		panic("onboarding workflow requires organization service")
	}
	if accounts == nil {
		panic("onboarding workflow requires account factory")
	}
	if workspaces == nil {
		panic("onboarding workflow requires workspace service")
	}
	if datadog == nil {
		panic("onboarding workflow requires datadog service")
	}

	return &Workflow{
		preferences: preferences,
		orgs:        orgs,
		accounts:    accounts,
		workspaces:  workspaces,
		datadog:     datadog,
		readiness:   readiness,
	}
}

// State returns current onboarding projection and next step.
func (w *Workflow) State(ctx context.Context) (State, error) {
	pref, err := w.preferences.Snapshot(ctx)
	if err != nil {
		return State{}, err
	}

	state, err := w.projectState(ctx, pref)
	if err != nil {
		return State{}, err
	}
	state.NextStep = nextStep(state)
	return state, nil
}

// Refresh is an alias for State and is useful for polling loops.
func (w *Workflow) Refresh(ctx context.Context) (State, error) {
	return w.State(ctx)
}

func (w *Workflow) projectState(ctx context.Context, pref preferences.Snapshot) (State, error) {
	state := State{Role: pref.Role}

	orgs, err := w.orgs.List(ctx)
	if err != nil {
		return state, err
	}
	state.Organizations = orgs
	state.SelectedOrganization = selectOrganization(orgs, pref.Organization)
	if state.SelectedOrganization == nil {
		state.DatadogDraft = w.currentDraft("")
		return state, nil
	}

	accounts, err := w.accounts(state.SelectedOrganization.ID).List(ctx)
	if err != nil {
		return state, err
	}
	state.Accounts = accounts
	state.SelectedAccount = selectAccount(accounts, pref.Account)
	if state.SelectedAccount == nil {
		state.DatadogDraft = w.currentDraft("")
		return state, nil
	}

	workspaces, err := w.workspaces.ListByAccount(ctx, state.SelectedAccount.ID)
	if err != nil {
		return state, err
	}
	state.Workspaces = workspaces
	state.SelectedWorkspace = selectWorkspace(workspaces, pref.Workspace)
	if state.SelectedWorkspace == nil {
		state.DatadogDraft = w.currentDraft(state.SelectedAccount.ID)
		return state, nil
	}

	ddAccount, err := w.datadog.GetByAccount(ctx, state.SelectedAccount.ID)
	if err != nil {
		return state, err
	}
	state.DatadogAccount = ddAccount
	if ddAccount != nil {
		status, err := w.datadog.Status(ctx, ddAccount.ID)
		if err != nil {
			return state, err
		}
		state.DatadogStatus = status
	}

	if w.readiness != nil {
		ready, err := w.readiness.Ready(ctx)
		if err != nil {
			return state, err
		}
		state.PowerSyncReady = ready
	} else {
		state.PowerSyncReady = true
	}

	state.DatadogDraft = w.currentDraft(state.SelectedAccount.ID)
	return state, nil
}

func (w *Workflow) currentStateWithError(ctx context.Context, err error) (State, error) {
	state, refreshErr := w.State(ctx)
	if refreshErr != nil {
		return State{}, err
	}
	return state, err
}
