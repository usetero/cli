package onboarding

import (
	"context"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/integrations/integrationstest"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/preferences/preferencestest"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/domains/tenancy/tenancytest"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
)

func newTestWorkflow(
	t *testing.T,
	pref preferences.Snapshot,
	orgs []tenancy.Organization,
	accounts map[tenancy.OrganizationID][]tenancy.Account,
	workspaces map[tenancy.AccountID][]tenancy.Workspace,
	datadogByAccount map[tenancy.AccountID]*integrations.DatadogAccount,
	datadogStatus map[integrations.DatadogAccountID]*integrations.DatadogStatus,
	ready bool,
) (*Workflow, *preferencestest.MockService, *integrationstest.MockDatadogService, *int) {
	t.Helper()

	snapshot := pref
	setScopeCalls := 0
	prefs := &preferencestest.MockService{
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
			setScopeCalls++
			snapshot.Organization = selection.OrganizationID
			snapshot.Account = selection.AccountID
			snapshot.Workspace = selection.WorkspaceID
			return nil
		},
	}

	datadog := &integrationstest.MockDatadogService{
		GetByAccountFn: func(_ context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
			return datadogByAccount[accountID], nil
		},
		ValidateAPIKeyFn: func(context.Context, integrations.DatadogAPIKeyValidation) (bool, string, error) {
			return true, "", nil
		},
		CreateFn: func(_ context.Context, input integrations.DatadogAccountCreate) (integrations.DatadogAccountID, error) {
			id := integrations.DatadogAccountID("dd_1")
			datadogByAccount[input.AccountID] = &integrations.DatadogAccount{
				ID:   id,
				Name: input.Name.String(),
				Site: input.Site,
			}
			if datadogStatus[id] == nil {
				datadogStatus[id] = &integrations.DatadogStatus{ReadyForUse: false}
			}
			return id, nil
		},
		StatusFn: func(_ context.Context, accountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
			return datadogStatus[accountID], nil
		},
	}

	workflow := NewWorkflow(
		prefs,
		&tenancytest.MockOrganizationService{
			ListFn: func(context.Context) ([]tenancy.Organization, error) { return orgs, nil },
			CreateFn: func(context.Context, tenancy.OrganizationCreate) (tenancy.OrganizationBootstrap, error) {
				return tenancy.OrganizationBootstrap{
					Organization: tenancy.Organization{ID: "org_1", Name: "Org 1"},
					Account:      tenancy.Account{ID: "acct_1", Name: "Account 1"},
					Workspace:    tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
				}, nil
			},
		},
		func(orgID tenancy.OrganizationID) tenancy.AccountService {
			return &tenancytest.MockAccountService{
				ListFn: func(context.Context) ([]tenancy.Account, error) { return accounts[orgID], nil },
				CreateFn: func(context.Context, tenancy.AccountCreate) (tenancy.AccountID, error) {
					return "acct_new", nil
				},
			}
		},
		&tenancytest.MockWorkspaceService{
			ListByAccountFn: func(_ context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
				return workspaces[accountID], nil
			},
		},
		datadog,
		syncertest.MockReadinessService{
			ReadyFn: func(context.Context) (bool, error) { return ready, nil },
		},
	)

	return workflow, prefs, datadog, &setScopeCalls
}

func TestWorkflow_CreateOrganizationAppliesBootstrapScope(t *testing.T) {
	t.Parallel()

	workflow, _, _, setScopeCalls := newTestWorkflow(
		t,
		preferences.Snapshot{Role: preferences.RoleEngineer},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		map[tenancy.AccountID]*integrations.DatadogAccount{},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
		false,
	)

	state, err := workflow.CreateOrganization(context.Background(), tenancy.OrganizationCreate{Name: "Org 1"})
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
	if state.NextStep != StepDatadogRegion {
		t.Fatalf("expected StepDatadogRegion, got %q", state.NextStep)
	}
	if *setScopeCalls != 1 {
		t.Fatalf("expected one SetScope call, got %d", *setScopeCalls)
	}
}

func TestWorkflow_IgnoresStaleOrganizationPreference(t *testing.T) {
	t.Parallel()

	workflow, _, _, _ := newTestWorkflow(
		t,
		preferences.Snapshot{Role: preferences.RoleEngineer, Organization: "org_stale"},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}, {ID: "org_2", Name: "Org 2"}},
		nil,
		nil,
		map[tenancy.AccountID]*integrations.DatadogAccount{},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
		false,
	)

	state, err := workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.SelectedOrganization != nil {
		t.Fatalf("expected stale organization preference to be ignored, got %+v", state.SelectedOrganization)
	}
	if state.NextStep != StepOrganizationSelect {
		t.Fatalf("expected StepOrganizationSelect, got %q", state.NextStep)
	}
}

func TestWorkflow_InvalidDatadogKeyReturnsErrorAndPreservesState(t *testing.T) {
	t.Parallel()

	workflow, _, datadog, _ := newTestWorkflow(
		t,
		preferences.Snapshot{
			Role:         preferences.RoleEngineer,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		map[tenancy.AccountID]*integrations.DatadogAccount{},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
		false,
	)
	datadog.ValidateAPIKeyFn = func(context.Context, integrations.DatadogAPIKeyValidation) (bool, string, error) {
		return false, "datadog rejected key", nil
	}

	state, err := workflow.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1)
	if err != nil {
		t.Fatalf("SetDatadogSite() error = %v", err)
	}
	if !state.DatadogDraft.Site.Valid() {
		t.Fatal("expected datadog site to be recorded")
	}

	state, err = workflow.SubmitDatadogAPIKey(context.Background(), integrations.DatadogAPIKeySubmission{
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
	if state.NextStep != StepDatadogAPIKey {
		t.Fatalf("expected StepDatadogAPIKey, got %q", state.NextStep)
	}
}

func TestWorkflow_PowerSyncReadinessGatesDone(t *testing.T) {
	t.Parallel()

	workflow, _, _, _ := newTestWorkflow(
		t,
		preferences.Snapshot{
			Role:         preferences.RoleEngineer,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		map[tenancy.AccountID]*integrations.DatadogAccount{
			"acct_1": {ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{
			"dd_1": {ReadyForUse: true},
		},
		false,
	)

	state, err := workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.NextStep != StepPowerSyncReady {
		t.Fatalf("expected StepPowerSyncReady, got %q", state.NextStep)
	}

	workflow, _, _, _ = newTestWorkflow(
		t,
		preferences.Snapshot{
			Role:         preferences.RoleEngineer,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		[]tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		map[tenancy.AccountID]*integrations.DatadogAccount{
			"acct_1": {ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		},
		map[integrations.DatadogAccountID]*integrations.DatadogStatus{
			"dd_1": {ReadyForUse: true},
		},
		true,
	)

	state, err = workflow.State(context.Background())
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}
	if state.NextStep != StepDone {
		t.Fatalf("expected StepDone, got %q", state.NextStep)
	}
}
