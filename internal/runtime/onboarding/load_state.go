package onboarding

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
)

func (s *Service) loadState(ctx context.Context, pref preferences.Snapshot) (State, error) {
	orgs, err := s.orgs.List(ctx)
	if err != nil {
		return State{}, err
	}

	state := State{Role: pref.Role, Organizations: Organizations(orgs)}
	state.SelectedOrganization = state.Organizations.Select(pref.Organization)
	if state.SelectedOrganization == nil {
		state.DatadogDraft = s.currentDraft("")
		return state, nil
	}

	accounts, err := s.accounts(state.SelectedOrganization.ID).List(ctx)
	if err != nil {
		return State{}, err
	}
	state.Accounts = Accounts(accounts)
	state.SelectedAccount = state.Accounts.Select(pref.Account)
	if state.SelectedAccount == nil {
		state.DatadogDraft = s.currentDraft("")
		return state, nil
	}

	workspaces, err := s.workspaces.ListByAccount(ctx, state.SelectedAccount.ID)
	if err != nil {
		return State{}, err
	}
	state.Workspaces = Workspaces(workspaces)
	state.SelectedWorkspace = state.Workspaces.Select(pref.Workspace)
	if state.SelectedWorkspace == nil {
		state.DatadogDraft = s.currentDraft(state.SelectedAccount.ID)
		return state, nil
	}

	ddAccount, err := s.datadog.GetByAccount(ctx, state.SelectedAccount.ID)
	if err != nil {
		return State{}, err
	}
	state.DatadogAccount = ddAccount
	if ddAccount != nil {
		status, err := s.datadog.Status(ctx, ddAccount.ID)
		if err != nil {
			return State{}, err
		}
		state.DatadogStatus = status
	}

	if s.readiness != nil {
		ready, err := s.readiness.Ready(ctx)
		if err != nil {
			return State{}, err
		}
		state.PowerSyncReady = ready
	} else {
		state.PowerSyncReady = true
	}

	state.DatadogDraft = s.currentDraft(state.SelectedAccount.ID)
	return state, nil
}
