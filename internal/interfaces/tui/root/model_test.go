package root

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	onboardingscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding"
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
}

func (s onboardingRuntimeStub) State(context.Context) (onboardingruntime.State, error) {
	if s.state.NextStep == "" {
		s.state.NextStep = onboardingruntime.StepRoleSelect
	}
	return s.state, s.err
}
func (onboardingRuntimeStub) SetRole(context.Context, preferences.RoleSelection) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) SelectOrganization(context.Context, preferences.OrganizationSelection) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) CreateOrganization(context.Context, tenancy.OrganizationCreate) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) SelectAccount(context.Context, preferences.AccountSelection) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) CreateAccount(context.Context, tenancy.AccountCreate) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) SelectWorkspace(context.Context, preferences.WorkspaceSelection) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) SetDatadogSite(context.Context, integrations.DatadogSite) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) SubmitDatadogAPIKey(context.Context, integrations.DatadogAPIKeySubmission) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}
func (onboardingRuntimeStub) SubmitDatadogAppKey(context.Context, integrations.DatadogAppKeySubmission) (onboardingruntime.State, error) {
	return onboardingruntime.State{}, nil
}

type sessionRuntimeStub struct{}

func (sessionRuntimeStub) Ensure(context.Context, sessionruntime.Scope) error { return nil }
func (sessionRuntimeStub) Status() sessionruntime.Status {
	return sessionruntime.Status{
		Sync: &pssyncer.Disconnected{},
	}
}
func (sessionRuntimeStub) SyncState() pssyncer.State {
	return &pssyncer.Disconnected{}
}

type mutableSessionStatusStub struct {
	status sessionruntime.Status
}

func (s *mutableSessionStatusStub) Status() sessionruntime.Status { return s.status }

func newOnboardingModel(t *testing.T) *onboardingscreen.Model {
	t.Helper()
	return newOnboardingModelWithRuntime(t, onboardingRuntimeStub{})
}

func newOnboardingModelWithRuntime(t *testing.T, runtime onboardingRuntimeStub) *onboardingscreen.Model {
	t.Helper()
	appTheme := theme.New(false)
	session := sessionRuntimeStub{}
	return onboardingscreen.New(
		runtime,
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

func TestModel_RendersOnboardingInFrame(t *testing.T) {
	appTheme := theme.New(false)
	model := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)
	msg := model.Init()()
	model.Update(msg)

	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "TERO") {
		t.Fatalf("expected framed app name, got %q", view)
	}
	if !strings.Contains(view, "Select your role:") {
		t.Fatalf("expected onboarding content, got %q", view)
	}
	if !strings.Contains(view, "ctrl+c") || !strings.Contains(view, "esc") {
		t.Fatalf("expected footer help content, got %q", view)
	}
}

func TestModel_CtrlCQuitsImmediately(t *testing.T) {
	appTheme := theme.New(false)
	model := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)
	updated, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	rootModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected *Model, got %T", updated)
	}
	if !rootModel.quit {
		t.Fatal("expected quit=true after ctrl+c")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestModel_EscapeOpensQuitDialog(t *testing.T) {
	appTheme := theme.New(false)
	model := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)

	updated, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	rootModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected *Model, got %T", updated)
	}
	if cmd != nil {
		t.Fatal("did not expect esc to quit immediately")
	}
	if rootModel.quitDialog == nil {
		t.Fatal("expected quit dialog to open")
	}

	view := ansi.Strip(rootModel.View().Content)
	if !strings.Contains(view, "Are you sure you want to quit?") {
		t.Fatalf("expected quit dialog content, got %q", view)
	}
	if !strings.Contains(view, "Yes") || !strings.Contains(view, "No") {
		t.Fatalf("expected quit dialog buttons, got %q", view)
	}
}

func TestModel_EscapeThenEnterCancelsByDefault(t *testing.T) {
	appTheme := theme.New(false)
	model := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)

	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	rootModel := updated.(*Model)
	updated, cmd := rootModel.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	rootModel = updated.(*Model)
	if cmd != nil {
		t.Fatal("did not expect enter on default No selection to quit")
	}
	if rootModel.quitDialog != nil {
		t.Fatal("expected quit dialog to close after cancel")
	}
}

func TestModel_QuitDialogConfirmFlow(t *testing.T) {
	appTheme := theme.New(false)
	model := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)

	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	rootModel := updated.(*Model)
	updated, _ = rootModel.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	rootModel = updated.(*Model)
	updated, cmd := rootModel.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	rootModel = updated.(*Model)
	if !rootModel.quit {
		t.Fatal("expected confirmed quit to mark model as quitting")
	}
	if cmd == nil {
		t.Fatal("expected confirmed quit command")
	}
}

func TestModel_ChromeWrapsLoadingAndErrorScreens(t *testing.T) {
	// Loading state before init response resolves.
	appTheme := theme.New(false)
	loadingModel := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)
	loadingView := ansi.Strip(loadingModel.View().Content)
	if !strings.Contains(loadingView, "Loading onboarding state...") {
		t.Fatalf("expected loading content in shell, got %q", loadingView)
	}

	// Error state after init response fails.
	errModel := New(logging.Scope{}, newOnboardingModelWithRuntime(t, onboardingRuntimeStub{
		err: context.DeadlineExceeded,
	}), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)
	msg := errModel.Init()()
	errModel.Update(msg)
	errView := ansi.Strip(errModel.View().Content)
	if !strings.Contains(errView, "Failed to load onboarding state.") {
		t.Fatalf("expected error content in shell, got %q", errView)
	}
	if !strings.Contains(errView, "╭") {
		t.Fatalf("expected error card border chrome, got %q", errView)
	}
}

func TestModel_StatusBarLifecycleProgression(t *testing.T) {
	appTheme := theme.New(false)
	session := &mutableSessionStatusStub{
		status: sessionruntime.Status{Running: false, Sync: &pssyncer.Disconnected{}},
	}
	model := New(
		logging.Scope{},
		newOnboardingModel(t),
		statusbar.New(session, "dev", appTheme),
		appTheme,
	)

	// Apply terminal size so status bar fitting logic is active.
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	offlineView := ansi.Strip(model.View().Content)
	if !strings.Contains(offlineView, "●") {
		t.Fatalf("expected offline sync dot in status bar, got %q", offlineView)
	}

	session.status = sessionruntime.Status{
		Running: true,
		Sync: &pssyncer.Syncing{
			Progress: &pssyncer.Progress{Downloaded: 3, Total: 10},
		},
	}
	syncingView := ansi.Strip(model.View().Content)
	if !strings.Contains(syncingView, "●") {
		t.Fatalf("expected syncing sync dot in status bar, got %q", syncingView)
	}

	session.status = sessionruntime.Status{Running: true, Sync: &pssyncer.Ready{}}
	readyView := ansi.Strip(model.View().Content)
	if !strings.Contains(readyView, "●") {
		t.Fatalf("expected ready sync dot in status bar, got %q", readyView)
	}
	if strings.Contains(readyView, "ready") || strings.Contains(readyView, "sync 3/10") || strings.Contains(readyView, "offline") {
		t.Fatalf("did not expect textual sync labels in status bar, got %q", readyView)
	}
}

func TestModel_ViewTerminalPolicy(t *testing.T) {
	appTheme := theme.New(false)
	model := New(logging.Scope{}, newOnboardingModel(t), statusbar.New(sessionRuntimeStub{}, "dev", appTheme), appTheme)
	msg := model.Init()()
	model.Update(msg)

	view := model.View()
	if !view.AltScreen {
		t.Fatal("expected AltScreen=true")
	}
	if view.MouseMode != tea.MouseModeNone {
		t.Fatalf("expected MouseModeNone, got %v", view.MouseMode)
	}
	if view.WindowTitle != "Tero" {
		t.Fatalf("expected window title Tero, got %q", view.WindowTitle)
	}
}
