package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/integrations/integrationstest"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/preferences/preferencestest"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/domains/tenancy/tenancytest"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
)

func newServiceForTest(pref preferences.Snapshot, orgs []tenancy.Organization, accounts map[tenancy.OrganizationID][]tenancy.Account, workspaces map[tenancy.AccountID][]tenancy.Workspace, dd map[tenancy.AccountID]*integrations.DatadogAccount, status map[integrations.DatadogAccountID]*integrations.DatadogStatus, ready bool) *Service {
	snapshot := pref

	preferencesService := &preferencestest.MockService{
		SnapshotFn: func(context.Context) (preferences.Snapshot, error) { return snapshot, nil },
		SetRoleFn: func(_ context.Context, selection preferences.RoleSelection) error {
			snapshot.Role = selection.Role
			return nil
		},
		SetOrganizationFn: func(_ context.Context, selection preferences.OrganizationSelection) error {
			snapshot.Organization = selection.OrganizationID
			snapshot.Account = ""
			snapshot.Workspace = ""
			return nil
		},
		SetAccountFn: func(_ context.Context, selection preferences.AccountSelection) error {
			snapshot.Account = selection.AccountID
			snapshot.Workspace = ""
			return nil
		},
		SetWorkspaceFn: func(_ context.Context, selection preferences.WorkspaceSelection) error {
			snapshot.Workspace = selection.WorkspaceID
			return nil
		},
		SetScopeFn: func(_ context.Context, selection preferences.ScopeSelection) error {
			snapshot.Organization = selection.OrganizationID
			snapshot.Account = selection.AccountID
			snapshot.Workspace = selection.WorkspaceID
			return nil
		},
		ClearScopeFn: func(context.Context) error {
			snapshot.Organization = ""
			snapshot.Account = ""
			snapshot.Workspace = ""
			return nil
		},
	}

	svc := NewService(
		preferencesService,
		&tenancytest.MockOrganizationService{
			ListFn: func(context.Context) ([]tenancy.Organization, error) {
				return orgs, nil
			},
		},
		func(organizationID tenancy.OrganizationID) tenancy.AccountService {
			return &tenancytest.MockAccountService{
				CreateFn: func(context.Context, tenancy.AccountCreate) (tenancy.AccountID, error) {
					return "acct_new", nil
				},
				DeleteFn: func(context.Context, tenancy.AccountID) error { return nil },
				ListFn: func(context.Context) ([]tenancy.Account, error) {
					return accounts[organizationID], nil
				},
			}
		},
		&tenancytest.MockWorkspaceService{
			CreateFn: func(context.Context, tenancy.WorkspaceCreate) (tenancy.WorkspaceID, error) {
				return "ws_new", nil
			},
			DeleteFn: func(context.Context, tenancy.WorkspaceID) error { return nil },
			ListByAccountFn: func(_ context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
				return workspaces[accountID], nil
			},
		},
		&integrationstest.MockDatadogService{
			GetByAccountFn: func(_ context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
				return dd[accountID], nil
			},
			ValidateAPIKeyFn: func(context.Context, integrations.DatadogAPIKeyValidation) (bool, string, error) {
				return true, "", nil
			},
			CreateFn: func(context.Context, integrations.DatadogAccountCreate) (integrations.DatadogAccountID, error) {
				return "", nil
			},
			StatusFn: func(_ context.Context, datadogAccountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
				return status[datadogAccountID], nil
			},
		},
		syncertest.MockReadinessService{
			ReadyFn: func(context.Context) (bool, error) { return ready, nil },
		},
	)
	return svc
}

func TestState_RoleFirst(t *testing.T) {
	svc := newServiceForTest(
		preferences.Snapshot{},
		[]tenancy.Organization{{ID: "org_1", Name: "Org"}},
		map[tenancy.OrganizationID][]tenancy.Account{"org_1": {{ID: "acct_1", Name: "A"}}},
		map[tenancy.AccountID][]tenancy.Workspace{"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W"}}},
		map[tenancy.AccountID]*integrations.DatadogAccount{"acct_1": {ID: "dd_1", Name: "DD", Site: integrations.DatadogSiteUS1}},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{"dd_1": {ReadyForUse: true}},
		true,
	)

	state, err := svc.State(context.Background())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.NextStep != StepRoleSelect {
		t.Fatalf("next step mismatch: got %q", state.NextStep)
	}
}

func TestState_OrganizationSelectionRules(t *testing.T) {
	orgs := []tenancy.Organization{{ID: "org_1", Name: "Org 1"}, {ID: "org_2", Name: "Org 2"}}
	accounts := map[tenancy.OrganizationID][]tenancy.Account{"org_2": {{ID: "acct_2", Name: "A2"}}}
	workspaces := map[tenancy.AccountID][]tenancy.Workspace{"acct_2": {{ID: "ws_2", AccountID: "acct_2", Name: "W2"}}}

	t.Run("stale preference ignored", func(t *testing.T) {
		svc := newServiceForTest(
			preferences.Snapshot{Role: preferences.RolePlatform, Organization: "org_stale"},
			orgs,
			accounts,
			workspaces,
			nil,
			nil,
			false,
		)
		state, err := svc.State(context.Background())
		if err != nil {
			t.Fatalf("state: %v", err)
		}
		if state.SelectedOrganization != nil {
			t.Fatalf("expected no selected org, got %+v", *state.SelectedOrganization)
		}
		if state.NextStep != StepOrganizationSelect {
			t.Fatalf("next step mismatch: got %q", state.NextStep)
		}
	})

	t.Run("valid preference auto-selects", func(t *testing.T) {
		svc := newServiceForTest(
			preferences.Snapshot{Role: preferences.RolePlatform, Organization: "org_2"},
			orgs,
			accounts,
			workspaces,
			nil,
			nil,
			false,
		)
		state, err := svc.State(context.Background())
		if err != nil {
			t.Fatalf("state: %v", err)
		}
		if state.SelectedOrganization == nil || state.SelectedOrganization.ID != "org_2" {
			t.Fatalf("expected selected org_2, got %+v", state.SelectedOrganization)
		}
	})
}

func TestMethods_SetRoleAndDatadogSite(t *testing.T) {
	svc := newServiceForTest(
		preferences.Snapshot{Organization: "org_1", Account: "acct_1", Workspace: "ws_1"},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{"org_1": {{ID: "acct_1", Name: "A1"}}},
		map[tenancy.AccountID][]tenancy.Workspace{"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W1"}}},
		nil,
		nil,
		false,
	)

	state, err := svc.SetRole(context.Background(), preferences.RoleSelection{Role: preferences.RoleEngineer})
	if err != nil {
		t.Fatalf("set role: %v", err)
	}
	if state.Role != preferences.RoleEngineer {
		t.Fatalf("role not persisted: %q", state.Role)
	}

	state, err = svc.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1)
	if err != nil {
		t.Fatalf("set site: %v", err)
	}
	if state.DatadogDraft.Site != integrations.DatadogSiteUS1 {
		t.Fatalf("draft site mismatch: %q", state.DatadogDraft.Site)
	}
	if state.NextStep != StepDatadogAPIKey {
		t.Fatalf("next step mismatch: got %q", state.NextStep)
	}
}

func TestState_DatadogAndPowerSyncOrdering(t *testing.T) {
	svc := newServiceForTest(
		preferences.Snapshot{Role: preferences.RolePlatform, Organization: "org_1", Account: "acct_1", Workspace: "ws_1"},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{"org_1": {{ID: "acct_1", Name: "A1"}}},
		map[tenancy.AccountID][]tenancy.Workspace{"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W1"}}},
		map[tenancy.AccountID]*integrations.DatadogAccount{"acct_1": {ID: "dd_1", Name: "DD", Site: integrations.DatadogSiteUS1}},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{"dd_1": {ReadyForUse: false}},
		false,
	)

	state, err := svc.State(context.Background())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.NextStep != StepDatadogDiscovery {
		t.Fatalf("expected datadog discovery step, got %q", state.NextStep)
	}

	svc = newServiceForTest(
		preferences.Snapshot{Role: preferences.RolePlatform, Organization: "org_1", Account: "acct_1", Workspace: "ws_1"},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{"org_1": {{ID: "acct_1", Name: "A1"}}},
		map[tenancy.AccountID][]tenancy.Workspace{"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W1"}}},
		map[tenancy.AccountID]*integrations.DatadogAccount{"acct_1": {ID: "dd_1", Name: "DD", Site: integrations.DatadogSiteUS1}},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{"dd_1": {ReadyForUse: true}},
		false,
	)
	state, err = svc.State(context.Background())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.NextStep != StepPowerSyncReady {
		t.Fatalf("expected powersync step, got %q", state.NextStep)
	}
}

func TestState_StaleAccountPreferenceRequiresSelection(t *testing.T) {
	svc := newServiceForTest(
		preferences.Snapshot{
			Role:         preferences.RoleEngineer,
			Organization: "org_1",
			Account:      "acct_stale",
		},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "A1"}, {ID: "acct_2", Name: "A2"}},
		},
		map[tenancy.AccountID][]tenancy.Workspace{},
		nil,
		nil,
		false,
	)

	state, err := svc.State(context.Background())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.SelectedAccount != nil {
		t.Fatalf("expected stale account preference to be ignored, got %+v", state.SelectedAccount)
	}
	if state.NextStep != StepAccountSelect {
		t.Fatalf("expected account select next step, got %q", state.NextStep)
	}
}

func TestState_StaleWorkspacePreferenceRequiresSelection(t *testing.T) {
	svc := newServiceForTest(
		preferences.Snapshot{
			Role:         preferences.RolePlatform,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_stale",
		},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "A1"}},
		},
		map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {
				{ID: "ws_1", AccountID: "acct_1", Name: "Main"},
				{ID: "ws_2", AccountID: "acct_1", Name: "Secondary"},
			},
		},
		nil,
		nil,
		false,
	)

	state, err := svc.State(context.Background())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.SelectedWorkspace != nil {
		t.Fatalf("expected stale workspace preference to be ignored, got %+v", state.SelectedWorkspace)
	}
	if state.NextStep != StepWorkspaceSelect {
		t.Fatalf("expected workspace select next step, got %q", state.NextStep)
	}
}

func TestState_PropagatesReadinessErrors(t *testing.T) {
	svc := NewService(
		&preferencestest.MockService{
			SnapshotFn: func(context.Context) (preferences.Snapshot, error) {
				return preferences.Snapshot{
					Role:         preferences.RolePlatform,
					Organization: "org_1",
					Account:      "acct_1",
					Workspace:    "ws_1",
				}, nil
			},
		},
		&tenancytest.MockOrganizationService{
			ListFn: func(context.Context) ([]tenancy.Organization, error) {
				return []tenancy.Organization{{ID: "org_1", Name: "Org 1"}}, nil
			},
		},
		func(organizationID tenancy.OrganizationID) tenancy.AccountService {
			return &tenancytest.MockAccountService{
				ListFn: func(context.Context) ([]tenancy.Account, error) {
					return []tenancy.Account{{ID: "acct_1", Name: "A1"}}, nil
				},
			}
		},
		&tenancytest.MockWorkspaceService{
			ListByAccountFn: func(_ context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
				return map[tenancy.AccountID][]tenancy.Workspace{
					"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W1"}},
				}[accountID], nil
			},
		},
		&integrationstest.MockDatadogService{
			GetByAccountFn: func(_ context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
				return map[tenancy.AccountID]*integrations.DatadogAccount{
					"acct_1": {ID: "dd_1", Name: "DD", Site: integrations.DatadogSiteUS1},
				}[accountID], nil
			},
			StatusFn: func(_ context.Context, datadogAccountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
				return map[integrations.DatadogAccountID]*integrations.DatadogStatus{
					"dd_1": {ReadyForUse: true},
				}[datadogAccountID], nil
			},
		},
		syncertest.MockReadinessService{
			ReadyFn: func(context.Context) (bool, error) { return false, errors.New("readiness unavailable") },
		},
	)

	_, err := svc.State(context.Background())
	if err == nil || err.Error() != "readiness unavailable" {
		t.Fatalf("expected readiness error passthrough, got %v", err)
	}
}
