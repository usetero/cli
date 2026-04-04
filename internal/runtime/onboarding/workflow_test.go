package onboarding_test

import (
	"context"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
	"github.com/usetero/cli/internal/runtime/onboardingtest"
)

func TestWorkflow_CreateOrganizationAppliesBootstrapScope(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot:      preferences.Snapshot{},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		Workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{},
		DatadogStatus:    map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
	})

	state, err := h.Workflow.CreateOrganization(context.Background(), tenancy.OrganizationCreate{Name: "Org 1"})
	if err != nil {
		t.Fatalf("CreateOrganization() error = %v", err)
	}
	if state.SelectedOrganization == nil || state.SelectedOrganization.ID != "org_1" {
		t.Fatalf("expected selected organization org_1, got %+v", state.SelectedOrganization)
	}
	if state.SelectedAccount == nil || state.SelectedAccount.ID != "acct_1" {
		t.Fatalf("expected selected account acct_1, got %+v", state.SelectedAccount)
	}
	if state.SelectedWorkspace == nil || state.SelectedWorkspace.ID != "ws_1" {
		t.Fatalf("expected selected workspace ws_1, got %+v", state.SelectedWorkspace)
	}
	if state.NextStep != runtimeonboarding.StepDatadogRegion {
		t.Fatalf("expected StepDatadogRegion, got %q", state.NextStep)
	}
	if h.PreferenceStore.SaveCalls != 1 {
		t.Fatalf("expected one preference save, got %d", h.PreferenceStore.SaveCalls)
	}
	if h.PreferenceStore.Snapshot.Organization != "org_1" || h.PreferenceStore.Snapshot.Account != "acct_1" || h.PreferenceStore.Snapshot.Workspace != "ws_1" {
		t.Fatalf("expected bootstrap scope to be stored, got %+v", h.PreferenceStore.Snapshot)
	}
}

func TestWorkflow_IgnoresStaleOrganizationPreference(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot:         preferences.Snapshot{Organization: "org_stale"},
		Organizations:    []tenancy.Organization{{ID: "org_1", Name: "Org 1"}, {ID: "org_2", Name: "Org 2"}},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{},
		DatadogStatus:    map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
	})

	state, err := h.Workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.SelectedOrganization != nil {
		t.Fatalf("expected stale organization preference to be ignored, got %+v", state.SelectedOrganization)
	}
	if state.NextStep != runtimeonboarding.StepOrganizationSelect {
		t.Fatalf("expected StepOrganizationSelect, got %q", state.NextStep)
	}
}

func TestWorkflow_IgnoresStaleAccountPreference(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_stale",
		},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {
				{ID: "acct_1", Name: "Account 1"},
				{ID: "acct_2", Name: "Account 2"},
			},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{},
		DatadogStatus:    map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
	})

	state, err := h.Workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.SelectedAccount != nil {
		t.Fatalf("expected stale account preference to be ignored, got %+v", state.SelectedAccount)
	}
	if state.NextStep != runtimeonboarding.StepAccountSelect {
		t.Fatalf("expected StepAccountSelect, got %q", state.NextStep)
	}
}

func TestWorkflow_IgnoresStaleWorkspacePreference(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_stale",
		},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		Workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {
				{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
				{ID: "ws_2", AccountID: "acct_1", Name: "Workspace 2"},
			},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{},
		DatadogStatus:    map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
	})

	state, err := h.Workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.SelectedWorkspace != nil {
		t.Fatalf("expected stale workspace preference to be ignored, got %+v", state.SelectedWorkspace)
	}
	if state.NextStep != runtimeonboarding.StepWorkspaceSelect {
		t.Fatalf("expected StepWorkspaceSelect, got %q", state.NextStep)
	}
}

func TestWorkflow_ExistingDatadogAccountSkipsSetup(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		Workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{
			"acct_1": {ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS5},
		},
		DatadogStatus: map[integrations.DatadogAccountID]*integrations.DatadogStatus{
			"dd_1": {ReadyForUse: false},
		},
	})

	state, err := h.Workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.DatadogAccount == nil || state.DatadogAccount.ID != "dd_1" {
		t.Fatalf("expected existing datadog account, got %+v", state.DatadogAccount)
	}
	if state.NextStep != runtimeonboarding.StepDatadogDiscovery {
		t.Fatalf("expected StepDatadogDiscovery, got %q", state.NextStep)
	}
}

func TestWorkflow_InvalidDatadogKeyReturnsErrorAndPreservesState(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		Workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{},
		DatadogStatus:    map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
	})
	h.Datadog.ValidateAPIKeyFn = func(context.Context, integrations.DatadogAPIKeyValidation) (bool, string, error) {
		return false, "datadog rejected key", nil
	}

	state, err := h.Workflow.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1)
	if err != nil {
		t.Fatalf("SetDatadogSite() error = %v", err)
	}
	if !state.DatadogDraft.Site.Valid() {
		t.Fatal("expected datadog site to be recorded")
	}

	state, err = h.Workflow.SubmitDatadogAPIKey(context.Background(), integrations.DatadogAPIKeySubmission{
		APIKey: integrations.DatadogAPIKey("bad"),
	})
	if err == nil || !strings.Contains(err.Error(), "datadog rejected key") {
		t.Fatalf("expected datadog validation error, got %v", err)
	}
	if state.DatadogDraft.Site != integrations.DatadogSiteUS1 {
		t.Fatalf("expected datadog site to remain set, got %q", state.DatadogDraft.Site)
	}
	if state.DatadogDraft.HasAPIKey {
		t.Fatal("did not expect API key to be marked valid")
	}
	if state.NextStep != runtimeonboarding.StepDatadogAPIKey {
		t.Fatalf("expected StepDatadogAPIKey, got %q", state.NextStep)
	}
}

func TestWorkflow_PowerSyncReadinessGatesDone(t *testing.T) {
	t.Parallel()

	h := onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		Workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{
			"acct_1": {ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		},
		DatadogStatus: map[integrations.DatadogAccountID]*integrations.DatadogStatus{
			"dd_1": {ReadyForUse: true},
		},
	})

	state, err := h.Workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.NextStep != runtimeonboarding.StepPowerSyncReady {
		t.Fatalf("expected StepPowerSyncReady, got %q", state.NextStep)
	}

	h = onboardingtest.NewHarness(t, onboardingtest.Config{
		Snapshot: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		Organizations: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		Accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		Workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		DatadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{
			"acct_1": {ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		},
		DatadogStatus: map[integrations.DatadogAccountID]*integrations.DatadogStatus{
			"dd_1": {ReadyForUse: true},
		},
		Ready: true,
	})

	state, err = h.Workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.NextStep != runtimeonboarding.StepDone {
		t.Fatalf("expected StepDone, got %q", state.NextStep)
	}
}
