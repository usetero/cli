package onboarding

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/runtime/onboarding/onboardingtest"
)

func newServiceForTest(pref preferences.Snapshot, orgs []tenancy.Organization, accounts map[tenancy.OrganizationID][]tenancy.Account, workspaces map[tenancy.AccountID][]tenancy.Workspace, dd map[tenancy.AccountID]*integrations.DatadogAccount, status map[integrations.DatadogAccountID]*integrations.DatadogStatus, ready bool) *Service {
	svc, err := NewService(
		&onboardingtest.PreferenceService{SnapshotValue: pref},
		&onboardingtest.OrganizationService{ListValue: orgs},
		func(organizationID tenancy.OrganizationID) tenancy.AccountService {
			return &onboardingtest.AccountService{ListValue: accounts[organizationID]}
		},
		&onboardingtest.WorkspaceService{ListByAccountValue: workspaces},
		&onboardingtest.DatadogService{ByAccountValue: dd, StatusValue: status},
		onboardingtest.ReadinessService{ReadyValue: ready},
	)
	if err != nil {
		panic(err)
	}
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

	state, err := svc.SetRole(context.Background(), preferences.RoleEngineer)
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
