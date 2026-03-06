package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	integrationsflow "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations"
	datadogapikey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/api_key"
	datadogappkey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/app_key"
	datadogregion "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/region"
	providerselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/provider/select"
	powersyncscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/powersync"
	"github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/role"
	tenancyflow "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy"
	accountcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/create"
	accountselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/select"
	organizationcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/create"
	organizationselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/select"
	workspaceselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/workspace/select"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	onboardingruntime "github.com/usetero/cli/internal/runtime/onboarding"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

type onboardingRuntimeStub struct {
	state onboardingruntime.State
	err   error

	setRoleState onboardingruntime.State
	setRoleErr   error
	setRoleCalls int

	selectOrganizationState onboardingruntime.State
	selectOrganizationErr   error
	selectOrganizationCalls int

	createOrganizationState onboardingruntime.State
	createOrganizationErr   error
	createOrganizationCalls int

	selectAccountState onboardingruntime.State
	selectAccountErr   error
	selectAccountCalls int

	createAccountState onboardingruntime.State
	createAccountErr   error
	createAccountCalls int

	selectWorkspaceState onboardingruntime.State
	selectWorkspaceErr   error
	selectWorkspaceCalls int

	setDatadogSiteState onboardingruntime.State
	setDatadogSiteErr   error
	setDatadogSiteCalls int

	submitDatadogAPIKeyState onboardingruntime.State
	submitDatadogAPIKeyErr   error
	submitDatadogAPIKeyCalls int

	submitDatadogAppKeyState onboardingruntime.State
	submitDatadogAppKeyErr   error
	submitDatadogAppKeyCalls int
}

func (s *onboardingRuntimeStub) State(context.Context) (onboardingruntime.State, error) {
	return s.state, s.err
}
func (s *onboardingRuntimeStub) SetRole(context.Context, preferences.RoleSelection) (onboardingruntime.State, error) {
	s.setRoleCalls++
	return s.setRoleState, s.setRoleErr
}
func (s *onboardingRuntimeStub) SelectOrganization(context.Context, preferences.OrganizationSelection) (onboardingruntime.State, error) {
	s.selectOrganizationCalls++
	return s.selectOrganizationState, s.selectOrganizationErr
}
func (s *onboardingRuntimeStub) CreateOrganization(context.Context, tenancy.OrganizationCreate) (onboardingruntime.State, error) {
	s.createOrganizationCalls++
	return s.createOrganizationState, s.createOrganizationErr
}
func (s *onboardingRuntimeStub) SelectAccount(context.Context, preferences.AccountSelection) (onboardingruntime.State, error) {
	s.selectAccountCalls++
	return s.selectAccountState, s.selectAccountErr
}
func (s *onboardingRuntimeStub) CreateAccount(context.Context, tenancy.AccountCreate) (onboardingruntime.State, error) {
	s.createAccountCalls++
	return s.createAccountState, s.createAccountErr
}
func (s *onboardingRuntimeStub) SelectWorkspace(context.Context, preferences.WorkspaceSelection) (onboardingruntime.State, error) {
	s.selectWorkspaceCalls++
	return s.selectWorkspaceState, s.selectWorkspaceErr
}
func (s *onboardingRuntimeStub) SetDatadogSite(context.Context, integrations.DatadogSite) (onboardingruntime.State, error) {
	s.setDatadogSiteCalls++
	return s.setDatadogSiteState, s.setDatadogSiteErr
}
func (s *onboardingRuntimeStub) SubmitDatadogAPIKey(context.Context, integrations.DatadogAPIKeySubmission) (onboardingruntime.State, error) {
	s.submitDatadogAPIKeyCalls++
	return s.submitDatadogAPIKeyState, s.submitDatadogAPIKeyErr
}
func (s *onboardingRuntimeStub) SubmitDatadogAppKey(context.Context, integrations.DatadogAppKeySubmission) (onboardingruntime.State, error) {
	s.submitDatadogAppKeyCalls++
	return s.submitDatadogAppKeyState, s.submitDatadogAppKeyErr
}

type sessionRuntimeStub struct {
	ensureCalls int
	ensureErr   error
	ensureScope []sessionruntime.Scope
	status      sessionruntime.Status
}

func (s *sessionRuntimeStub) Ensure(_ context.Context, scope sessionruntime.Scope) error {
	s.ensureCalls++
	s.ensureScope = append(s.ensureScope, scope)
	return s.ensureErr
}

func (s *sessionRuntimeStub) Status() sessionruntime.Status {
	if s.status.Sync == nil {
		s.status.Sync = &pssyncer.Disconnected{}
	}
	return s.status
}

func newModel(rt *onboardingRuntimeStub) *Model {
	return newModelWithSession(rt, &sessionRuntimeStub{})
}

func newModelWithSession(rt *onboardingRuntimeStub, session *sessionRuntimeStub) *Model {
	appTheme := theme.New(false)
	return New(
		rt,
		session,
		role.New(logging.Scope{}, appTheme),
		tenancyflow.New(
			organizationselect.New(logging.Scope{}, appTheme),
			organizationcreate.New(logging.Scope{}, appTheme),
			accountselect.New(logging.Scope{}, appTheme),
			accountcreate.New(logging.Scope{}, appTheme),
			workspaceselect.New(logging.Scope{}, appTheme),
			appTheme,
		),
		integrationsflow.New(
			providerselect.New(logging.Scope{}, appTheme),
			datadogregion.New(logging.Scope{}, appTheme),
			datadogapikey.New(logging.Scope{}, appTheme),
			datadogappkey.New(logging.Scope{}, appTheme),
			[]integrations.Provider{integrations.ProviderDatadog},
			appTheme,
		),
		powersyncscreen.New(session, appTheme),
		appTheme,
	)
}

func drive(t *testing.T, m *Model, msg tea.Msg) *Model {
	t.Helper()
	updated, _ := m.Update(msg)
	model, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected *Model, got %T", updated)
	}
	return model
}

func updateWithCmd(t *testing.T, m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	model, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected *Model, got %T", updated)
	}
	return model, cmd
}

func TestModel_LoadAndRoleRoute(t *testing.T) {
	m := newModel(&onboardingRuntimeStub{state: onboardingruntime.State{NextStep: onboardingruntime.StepRoleSelect}})
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected init command")
	}
	m = drive(t, m, cmd())
	if !strings.Contains(m.View().Content, "Select your role:") {
		t.Fatalf("expected role view, got %q", m.View().Content)
	}
}

func TestModel_OrganizationCreateRoute(t *testing.T) {
	m := newModel(&onboardingRuntimeStub{state: onboardingruntime.State{NextStep: onboardingruntime.StepOrganizationCreate}})
	m = drive(t, m, m.Init()())
	if !strings.Contains(m.View().Content, "Create your organization:") {
		t.Fatalf("expected organization create view, got %q", m.View().Content)
	}
}

func TestModel_AccountCreateRoute(t *testing.T) {
	m := newModel(&onboardingRuntimeStub{state: onboardingruntime.State{NextStep: onboardingruntime.StepAccountCreate}})
	m = drive(t, m, m.Init()())
	if !strings.Contains(m.View().Content, "Create your account:") {
		t.Fatalf("expected account create view, got %q", m.View().Content)
	}
}

func TestModel_WorkspaceSelectRoute(t *testing.T) {
	m := newModel(&onboardingRuntimeStub{
		state: onboardingruntime.State{
			NextStep: onboardingruntime.StepWorkspaceSelect,
			Workspaces: []tenancy.Workspace{
				{ID: "ws_1", Name: "One"},
			},
		},
	})
	m = drive(t, m, m.Init()())
	if !strings.Contains(m.View().Content, "Select your workspace:") {
		t.Fatalf("expected workspace select view, got %q", m.View().Content)
	}
}

func TestModel_DatadogRegionToAPIKeyFlow(t *testing.T) {
	rt := &onboardingRuntimeStub{
		state:               onboardingruntime.State{NextStep: onboardingruntime.StepDatadogRegion},
		setDatadogSiteState: onboardingruntime.State{NextStep: onboardingruntime.StepDatadogAPIKey},
	}
	m := newModel(rt)
	m = drive(t, m, m.Init()())
	_, cmd := m.Update(integrationsflow.SetDatadogSiteMsg{Site: integrations.DatadogSiteUS1})
	if cmd == nil {
		t.Fatal("expected set site command")
	}
	m = drive(t, m, cmd())
	if rt.setDatadogSiteCalls != 1 {
		t.Fatalf("expected setDatadogSite call, got %d", rt.setDatadogSiteCalls)
	}
	if !strings.Contains(m.View().Content, "Enter Datadog API key:") {
		t.Fatalf("expected datadog api key view, got %q", m.View().Content)
	}
}

func TestModel_DatadogAPIKeyErrorShowsInline(t *testing.T) {
	rt := &onboardingRuntimeStub{
		state:                  onboardingruntime.State{NextStep: onboardingruntime.StepDatadogAPIKey},
		submitDatadogAPIKeyErr: errors.New("invalid api key"),
	}
	m := newModel(rt)
	m = drive(t, m, m.Init()())
	_, cmd := m.Update(integrationsflow.SubmitDatadogAPIKeyMsg{
		Submission: integrations.DatadogAPIKeySubmission{APIKey: "x"},
	})
	if cmd == nil {
		t.Fatal("expected submit api key command")
	}
	m = drive(t, m, cmd())
	view := m.View().Content
	if !strings.Contains(view, "Enter Datadog API key:") || !strings.Contains(view, "invalid api key") {
		t.Fatalf("expected inline error on api key view, got %q", view)
	}
}

func TestModel_DatadogAppKeyToDiscoveryAndRefresh(t *testing.T) {
	rt := &onboardingRuntimeStub{
		state:                    onboardingruntime.State{NextStep: onboardingruntime.StepDatadogAppKey},
		submitDatadogAppKeyState: onboardingruntime.State{NextStep: onboardingruntime.StepDatadogDiscovery},
	}
	m := newModel(rt)
	m = drive(t, m, m.Init()())
	_, cmd := m.Update(integrationsflow.SubmitDatadogAppKeyMsg{
		Submission: integrations.DatadogAppKeySubmission{Name: "A", AppKey: "K"},
	})
	if cmd == nil {
		t.Fatal("expected submit app key command")
	}
	m = drive(t, m, cmd())
	if !strings.Contains(m.View().Content, "Waiting for Datadog discovery") {
		t.Fatalf("expected discovery view, got %q", m.View().Content)
	}

	// Discovery refresh applies new state from runtime polling.
	m = drive(t, m, stateResolvedMsg{
		state: onboardingruntime.State{NextStep: onboardingruntime.StepPowerSyncReady},
	})
	if !strings.Contains(m.View().Content, "Syncing your account data") {
		t.Fatalf("expected powersync view, got %q", m.View().Content)
	}
}

func TestModel_PowerSyncRefreshToDone(t *testing.T) {
	rt := &onboardingRuntimeStub{
		state: onboardingruntime.State{NextStep: onboardingruntime.StepPowerSyncReady},
	}
	m := newModel(rt)
	m = drive(t, m, m.Init()())
	m = drive(t, m, stateResolvedMsg{
		state: onboardingruntime.State{NextStep: onboardingruntime.StepDone},
	})
	if !strings.Contains(m.View().Content, "Welcome to Tero") {
		t.Fatalf("expected done view, got %q", m.View().Content)
	}
}

func TestModel_EnsureOnlyWhenScopeReadyAndPollTransitionsToDone(t *testing.T) {
	session := &sessionRuntimeStub{
		status: sessionruntime.Status{
			Sync: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 4, Total: 10}},
		},
	}
	rt := &onboardingRuntimeStub{}
	m := newModelWithSession(rt, session)

	org := tenancy.Organization{ID: "org_1", Name: "Org 1"}
	account := tenancy.Account{ID: "acc_1", Name: "Account 1"}

	noAccountState := onboardingruntime.State{
		NextStep:             onboardingruntime.StepAccountSelect,
		SelectedOrganization: &org,
	}
	withAccountState := onboardingruntime.State{
		NextStep:             onboardingruntime.StepPowerSyncReady,
		SelectedOrganization: &org,
		SelectedAccount:      &account,
	}

	// No account selected yet: session Ensure must not run.
	var cmd tea.Cmd
	m, cmd = updateWithCmd(t, m, stateResolvedMsg{state: noAccountState})
	if cmd != nil {
		if _, ok := cmd().(sessionEnsuredMsg); ok {
			t.Fatalf("did not expect Ensure command before account selection")
		}
	}
	if session.ensureCalls != 0 {
		t.Fatalf("expected 0 ensure calls before account selection, got %d", session.ensureCalls)
	}

	// Scope is ready: Ensure runs with org+account, and we land on PowerSync screen.
	m, cmd = updateWithCmd(t, m, stateResolvedMsg{state: withAccountState})
	if cmd == nil {
		t.Fatal("expected batched command in powersync step")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok || len(batch) == 0 {
		t.Fatalf("expected batch command, got %T", cmd())
	}
	_, _ = updateWithCmd(t, m, batch[0]()) // run Ensure command only; skip timer commands

	if session.ensureCalls != 1 {
		t.Fatalf("expected 1 ensure call, got %d", session.ensureCalls)
	}
	if len(session.ensureScope) != 1 || session.ensureScope[0].OrganizationID != "org_1" || session.ensureScope[0].AccountID != "acc_1" {
		t.Fatalf("unexpected ensure scope: %+v", session.ensureScope)
	}
	if !strings.Contains(m.View().Content, "Syncing your account data") {
		t.Fatalf("expected powersync view, got %q", m.View().Content)
	}

	// Poll tick triggers runtime refresh; refreshed state resolves to done.
	rt.state = onboardingruntime.State{NextStep: onboardingruntime.StepDone}
	_, cmd = updateWithCmd(t, m, pollStateTickMsg{})
	if cmd == nil {
		t.Fatal("expected polling batch command")
	}
	pollBatch, ok := cmd().(tea.BatchMsg)
	if !ok || len(pollBatch) == 0 {
		t.Fatalf("expected poll batch command, got %T", cmd())
	}
	m, _ = updateWithCmd(t, m, pollBatch[0]()) // loadStateCmd result

	if !strings.Contains(m.View().Content, "Welcome to Tero") {
		t.Fatalf("expected done view after poll refresh, got %q", m.View().Content)
	}
}
