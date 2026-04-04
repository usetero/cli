package onboarding

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/identity/identitytest"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
	"github.com/usetero/cli/internal/runtime/onboardingtest"
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
	if input := m.Input(); input == nil || input.Kind != core.InputConfirm {
		t.Fatalf("expected auth action input, got %#v", input)
	}
}

func TestModelInit_AuthenticatedLoadsInitialWorkflowState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{
		Snapshot: preferences.Snapshot{},
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
			"dd_1": {ReadyForUse: false},
		},
		Ready: true,
	}

	m := newTestModel(t, cfg, true)
	m.account = accountruntime.Status{HasCompletedInitialSync: true}
	m.state = runtimeonboarding.State{
		SelectedOrganization: &tenancy.Organization{ID: "org_1", Name: "Org 1"},
		SelectedAccount:      &tenancy.Account{ID: "acct_1", Name: "Account 1"},
		SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		DatadogAccount:       cfg.DatadogByAccount["acct_1"],
		DatadogStatus:        cfg.DatadogStatus["dd_1"],
		PowerSyncReady:       true,
		NextStep:             runtimeonboarding.StepDatadogDiscovery,
	}
	m.showDatadogDiscovery()

	cfg.DatadogStatus["dd_1"].ReadyForUse = true

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

func TestModelAccountRuntimeStatus_ExposesPowerSyncFailureAsShellError(t *testing.T) {
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
			Scope: accountruntime.Scope{
				Organization: tenancy.Organization{ID: "org_1", Name: "Org 1"},
				Account:      tenancy.Account{ID: "acct_1", Name: "Account 1"},
				Workspace:    tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
			},
			Sync: &pssyncer.Error{Err: errors.New("boom")},
		},
	})

	if m.Active() != m.psReady {
		t.Fatalf("expected powersync step to remain active on failure")
	}
	if busy := m.Busy(); busy != nil {
		t.Fatalf("expected no shell busy state on failure, got %#v", busy)
	}
	if err := m.Error(); err == nil || err.Message != "Failed to prepare your workspace." || err.Detail != "boom" {
		t.Fatalf("unexpected shell error state: %#v", err)
	}
}

func TestModelProviders_ReflectLoadingAndBusyState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	m.loading = true

	if busy := m.Busy(); busy == nil || busy.Label != "Loading Onboarding State" {
		t.Fatalf("expected loading busy state, got %#v", busy)
	}
	if input := m.Input(); input != nil {
		t.Fatalf("expected no input while loading, got %#v", input)
	}
	if help := m.ShortHelp(); help != nil {
		t.Fatalf("expected no help while loading, got %+v", help)
	}

	m.loading = false
	m.busy = &core.Busy{Label: "Selecting Account"}

	if busy := m.Busy(); busy == nil || busy.Label != "Selecting Account" {
		t.Fatalf("expected explicit busy state, got %#v", busy)
	}
	if input := m.Input(); input != nil {
		t.Fatalf("expected no input while busy, got %#v", input)
	}
	if help := m.ShortHelp(); help != nil {
		t.Fatalf("expected no help while busy, got %+v", help)
	}
}

func TestModelStateLoaded_NotAuthenticatedRoutesBackToAuth(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, workflowConfig{}, true)
	m.showOrganizationCreate()

	_, cmd := m.Update(stateLoadedMsg{Err: identity.ErrNotAuthenticated})
	if cmd != nil {
		t.Fatalf("expected no follow-up command when routing back to auth")
	}
	if m.Active() != m.auth {
		t.Fatalf("expected auth step to become active")
	}
	if m.loadErr != nil {
		t.Fatalf("expected load error to be cleared, got %v", m.loadErr)
	}
	if m.notice == nil || m.notice.Message == "" {
		t.Fatalf("expected a session-ended notice")
	}
}

type workflowConfig = onboardingtest.Config

func newTestModel(t *testing.T, cfg workflowConfig, authenticated bool) *Model {
	t.Helper()

	h := onboardingtest.NewHarness(t, onboardingtest.Config(cfg))
	return New(logging.Scope{}, newIdentityService(authenticated), h.Workflow, theme.New(false))
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
