package onboarding

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/identity/identitytest"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/integrations/integrationstest"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/preferences/preferencestest"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/domains/tenancy/tenancytest"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
)

func TestModelInit_UnauthenticatedShowsAuthStep(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, false)

	cmd := m.Init()
	if cmd != nil {
		t.Fatalf("expected no load command before authentication")
	}
	if m.Active() != m.auth {
		t.Fatalf("expected auth step to be active")
	}
	if input := m.Input(); input == nil || input.Kind != core.InputAction {
		t.Fatalf("expected auth action input, got %#v", input)
	}
}

func TestModelInit_AuthenticatedLoadsInitialWorkflowState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{
		pref: preferences.Snapshot{},
	}, true)

	cmd := m.Init()
	if cmd == nil {
		t.Fatalf("expected authenticated init to load state")
	}

	msg := cmd()
	stateMsg, ok := msg.(stateLoadedMsg)
	if !ok {
		t.Fatalf("expected stateLoadedMsg, got %#v", msg)
	}

	_, _ = m.Update(stateMsg)
	if m.Active() != m.orgCreate {
		t.Fatalf("expected organization create step after initial state load")
	}
}

func TestModelStateLoaded_PublishesSelectedScopeMessages(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	state := runtimeonboarding.State{
		SelectedOrganization: &tenancy.Organization{ID: "org_1", Name: "Org 1"},
		SelectedAccount:      &tenancy.Account{ID: "acct_1", Name: "Account 1"},
		SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		DatadogAccount:       &integrations.DatadogAccount{ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		DatadogStatus:        &integrations.DatadogStatus{ReadyForUse: true},
		PowerSyncReady:       true,
		NextStep:             runtimeonboarding.StepDone,
	}

	_, cmd := m.Update(stateLoadedMsg{State: state})
	msgs := runCmdMessages(cmd)

	var (
		gotOrg     events.OrganizationSelectedMsg
		gotAccount events.AccountSelectedMsg
	)
	foundOrg := false
	foundAccount := false

	for _, msg := range msgs {
		switch typed := msg.(type) {
		case events.OrganizationSelectedMsg:
			foundOrg = true
			gotOrg = typed
		case events.AccountSelectedMsg:
			foundAccount = true
			gotAccount = typed
		}
	}

	if !foundOrg {
		t.Fatalf("expected organization-selected message")
	}
	if gotOrg.Organization.ID != "org_1" {
		t.Fatalf("unexpected organization message: %+v", gotOrg)
	}
	if !foundAccount {
		t.Fatalf("expected account-selected message")
	}
	if gotAccount.Scope.Account.ID != "acct_1" || gotAccount.Scope.Workspace.ID != "ws_1" {
		t.Fatalf("unexpected account scope message: %+v", gotAccount.Scope)
	}
}

func TestModelStateLoaded_RoutesToDatadogDiscoveryAndSchedulesRefresh(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	state := runtimeonboarding.State{
		SelectedOrganization: &tenancy.Organization{ID: "org_1", Name: "Org 1"},
		SelectedAccount:      &tenancy.Account{ID: "acct_1", Name: "Account 1"},
		SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		DatadogAccount:       &integrations.DatadogAccount{ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		DatadogStatus:        &integrations.DatadogStatus{ReadyForUse: false, EventCount: 100, AnalyzedCount: 42},
		NextStep:             runtimeonboarding.StepDatadogDiscovery,
	}

	_, cmd := m.Update(stateLoadedMsg{State: state})
	if m.Active() != m.ddDiscovery {
		t.Fatalf("expected Datadog discovery step to be active")
	}
	if input := m.Input(); input == nil || input.Label == "" {
		t.Fatalf("expected discovery input label, got %#v", input)
	}

	msgs := runCmdMessages(cmd)
	if len(msgs) == 0 {
		t.Fatalf("expected discovery refresh command")
	}
	foundRefresh := false
	for _, msg := range msgs {
		if _, ok := msg.(refreshRequestedMsg); ok {
			foundRefresh = true
			break
		}
	}
	if !foundRefresh {
		t.Fatalf("expected refreshRequestedMsg in routed commands, got %#v", msgs)
	}
}

func TestModelRefreshRequested_IgnoresNonDiscoverySteps(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	m.showOrganizationCreate()

	_, cmd := m.Update(refreshRequestedMsg{})
	if cmd != nil {
		t.Fatalf("expected no refresh command outside discovery")
	}
}

func TestModelRefreshRequested_ReprojectsDiscoveryState(t *testing.T) {
	t.Parallel()

	cfg := workflowConfig{
		pref: preferences.Snapshot{
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		},
		orgs: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		accounts: map[tenancy.OrganizationID][]tenancy.Account{
			"org_1": {{ID: "acct_1", Name: "Account 1"}},
		},
		workspaces: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		},
		datadogByAccount: map[tenancy.AccountID]*integrations.DatadogAccount{
			"acct_1": {ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		},
		datadogStatus: map[integrations.DatadogAccountID]*integrations.DatadogStatus{
			"dd_1": {ReadyForUse: false},
		},
		ready: true,
	}

	m := newTestModel(t, cfg, true)
	m.account = accountruntime.Status{HasCompletedInitialSync: true}
	m.state = runtimeonboarding.State{
		SelectedOrganization: &tenancy.Organization{ID: "org_1", Name: "Org 1"},
		SelectedAccount:      &tenancy.Account{ID: "acct_1", Name: "Account 1"},
		SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		DatadogAccount:       cfg.datadogByAccount["acct_1"],
		DatadogStatus:        cfg.datadogStatus["dd_1"],
		PowerSyncReady:       true,
		NextStep:             runtimeonboarding.StepDatadogDiscovery,
	}
	m.showDatadogDiscovery()

	cfg.datadogStatus["dd_1"].ReadyForUse = true

	_, cmd := m.Update(refreshRequestedMsg{})
	if cmd == nil {
		t.Fatalf("expected refresh to load state")
	}

	msgs := runCmdMessages(cmd)
	if len(msgs) != 1 {
		t.Fatalf("expected one state reload message, got %#v", msgs)
	}
	stateMsg, ok := msgs[0].(stateLoadedMsg)
	if !ok {
		t.Fatalf("expected stateLoadedMsg, got %#v", msgs[0])
	}

	_, _ = m.Update(stateMsg)
	if m.Active() != nil {
		t.Fatalf("expected discovery to finish once workflow reprojects to done")
	}
}

func TestModelAccountRuntimeStatus_ControlsFinalPowerSyncGate(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	m.state = runtimeonboarding.State{
		SelectedOrganization: &tenancy.Organization{ID: "org_1", Name: "Org 1"},
		SelectedAccount:      &tenancy.Account{ID: "acct_1", Name: "Account 1"},
		SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		DatadogAccount:       &integrations.DatadogAccount{ID: "dd_1", Name: "Datadog", Site: integrations.DatadogSiteUS1},
		DatadogStatus:        &integrations.DatadogStatus{ReadyForUse: true},
		PowerSyncReady:       true,
		NextStep:             runtimeonboarding.StepDone,
	}

	_, _ = m.Update(events.AccountRuntimeUpdatedMsg{
		Status: accountruntime.Status{
			Running:                 true,
			HasCompletedInitialSync: false,
		},
	})
	if m.Active() != m.psReady {
		t.Fatalf("expected cold account to stay on powersync gate")
	}

	_, _ = m.Update(events.AccountRuntimeUpdatedMsg{
		Status: accountruntime.Status{
			Running:                 true,
			HasCompletedInitialSync: true,
		},
	})
	if m.Active() != nil {
		t.Fatalf("expected warm account to clear onboarding gate")
	}
}

func TestModelProviders_ReflectLoadingAndBusyState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	m.loading = true

	if busy := m.Busy(); busy == nil || busy.Label != "Loading Onboarding State" {
		t.Fatalf("expected loading busy state, got %#v", busy)
	}
	if input := m.Input(); input == nil || input.Label != "Loading onboarding..." {
		t.Fatalf("expected loading input placeholder, got %#v", input)
	}
	if help := m.ShortHelp(); help != nil {
		t.Fatalf("expected no help while loading, got %+v", help)
	}

	m.loading = false
	m.busy = &core.Busy{Label: "Selecting Account"}

	if busy := m.Busy(); busy == nil || busy.Label != "Selecting Account" {
		t.Fatalf("expected explicit busy state, got %#v", busy)
	}
	if input := m.Input(); input == nil || input.Label != "Loading onboarding..." {
		t.Fatalf("expected busy override input, got %#v", input)
	}
	if help := m.ShortHelp(); help != nil {
		t.Fatalf("expected no help while busy, got %+v", help)
	}
}

type workflowConfig struct {
	pref             preferences.Snapshot
	orgs             []tenancy.Organization
	accounts         map[tenancy.OrganizationID][]tenancy.Account
	workspaces       map[tenancy.AccountID][]tenancy.Workspace
	datadogByAccount map[tenancy.AccountID]*integrations.DatadogAccount
	datadogStatus    map[integrations.DatadogAccountID]*integrations.DatadogStatus
	ready            bool
}

func newTestModel(t *testing.T, cfg workflowConfig, authenticated bool) *Model {
	t.Helper()

	if cfg.accounts == nil {
		cfg.accounts = map[tenancy.OrganizationID][]tenancy.Account{}
	}
	if cfg.workspaces == nil {
		cfg.workspaces = map[tenancy.AccountID][]tenancy.Workspace{}
	}
	if cfg.datadogByAccount == nil {
		cfg.datadogByAccount = map[tenancy.AccountID]*integrations.DatadogAccount{}
	}
	if cfg.datadogStatus == nil {
		cfg.datadogStatus = map[integrations.DatadogAccountID]*integrations.DatadogStatus{}
	}

	prefs := &preferencestest.MockService{
		SnapshotFn:        func(context.Context) (preferences.Snapshot, error) { return cfg.pref, nil },
		SetRoleFn:         func(context.Context, preferences.RoleSelection) error { return nil },
		SetOrganizationFn: func(context.Context, preferences.OrganizationSelection) error { return nil },
		SetAccountFn:      func(context.Context, preferences.AccountSelection) error { return nil },
		SetWorkspaceFn:    func(context.Context, preferences.WorkspaceSelection) error { return nil },
		SetScopeFn:        func(context.Context, preferences.ScopeSelection) error { return nil },
	}

	datadog := &integrationstest.MockDatadogService{
		GetByAccountFn: func(_ context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
			return cfg.datadogByAccount[accountID], nil
		},
		ValidateAPIKeyFn: func(context.Context, integrations.DatadogAPIKeyValidation) (bool, string, error) {
			return true, "", nil
		},
		CreateFn: func(context.Context, integrations.DatadogAccountCreate) (integrations.DatadogAccountID, error) {
			return "dd_1", nil
		},
		StatusFn: func(_ context.Context, accountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
			return cfg.datadogStatus[accountID], nil
		},
	}

	workflow := runtimeonboarding.NewWorkflow(
		prefs,
		&tenancytest.MockOrganizationService{
			ListFn: func(context.Context) ([]tenancy.Organization, error) { return cfg.orgs, nil },
		},
		func(orgID tenancy.OrganizationID) tenancy.AccountService {
			return &tenancytest.MockAccountService{
				ListFn: func(context.Context) ([]tenancy.Account, error) { return cfg.accounts[orgID], nil },
			}
		},
		&tenancytest.MockWorkspaceService{
			ListByAccountFn: func(_ context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
				return cfg.workspaces[accountID], nil
			},
		},
		datadog,
		syncertest.MockReadinessService{
			ReadyFn: func(context.Context) (bool, error) { return cfg.ready, nil },
		},
	)

	return New(logging.Scope{}, newIdentityService(authenticated), workflow, theme.New(false))
}

func newIdentityService(authenticated bool) *identity.Service {
	store := identitytest.NewTokenStore()
	if authenticated {
		store.AccessTokenValue = identity.AccessToken("access-token")
	}
	return identity.NewService(identitytest.NewProvider(), store, identity.NopLogger{})
}

func runCmdMessages(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	if msg == nil {
		return nil
	}

	switch typed := msg.(type) {
	case tea.BatchMsg:
		msgs := make([]tea.Msg, 0, len(typed))
		for _, child := range typed {
			msgs = append(msgs, runCmdMessages(child)...)
		}
		return msgs
	default:
		return []tea.Msg{typed}
	}
}
