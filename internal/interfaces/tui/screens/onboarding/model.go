package onboarding

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	integrationsflow "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations"
	powersyncscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/powersync"
	"github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/role"
	tenancyflow "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	onboardingruntime "github.com/usetero/cli/internal/runtime/onboarding"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

type route int

const (
	routeLoading route = iota
	routeRole
	routeTenancy
	routeIntegrations
	routePowerSyncReady
	routeDone
	routePlaceholder
	routeError
)

type childID string

const (
	childRole         childID = "role"
	childTenancy      childID = "tenancy"
	childIntegrations childID = "integrations"
	childPowerSync    childID = "powersync"
)

const powersyncPollInterval = 500 * time.Millisecond

type Runtime interface {
	State(ctx context.Context) (onboardingruntime.State, error)
	SetRole(ctx context.Context, selection preferences.RoleSelection) (onboardingruntime.State, error)
	SelectOrganization(ctx context.Context, selection preferences.OrganizationSelection) (onboardingruntime.State, error)
	CreateOrganization(ctx context.Context, create tenancy.OrganizationCreate) (onboardingruntime.State, error)
	SelectAccount(ctx context.Context, selection preferences.AccountSelection) (onboardingruntime.State, error)
	CreateAccount(ctx context.Context, create tenancy.AccountCreate) (onboardingruntime.State, error)
	SelectWorkspace(ctx context.Context, selection preferences.WorkspaceSelection) (onboardingruntime.State, error)
	SetDatadogSite(ctx context.Context, site integrations.DatadogSite) (onboardingruntime.State, error)
	SubmitDatadogAPIKey(ctx context.Context, submission integrations.DatadogAPIKeySubmission) (onboardingruntime.State, error)
	SubmitDatadogAppKey(ctx context.Context, submission integrations.DatadogAppKeySubmission) (onboardingruntime.State, error)
}

type Session interface {
	Ensure(ctx context.Context, scope sessionruntime.Scope) error
	Status() sessionruntime.Status
}

type stateResolvedMsg struct {
	state onboardingruntime.State
	err   error
}

type sessionEnsuredMsg struct {
	err error
}

type pollStateTickMsg struct{}

// Model owns onboarding flow orchestration and delegates phase behavior.
type Model struct {
	route route

	theme   theme.Theme
	runtime Runtime
	session Session
	step    onboardingruntime.Step
	loadErr error
	router  screen.Router[childID]

	role         *role.Model
	tenancy      *tenancyflow.Model
	integrations *integrationsflow.Model
	powersync    *powersyncscreen.Model
}

// New constructs the onboarding flow model.
func New(
	runtime Runtime,
	session Session,
	roleModel *role.Model,
	tenancyModel *tenancyflow.Model,
	integrationsModel *integrationsflow.Model,
	powersyncModel *powersyncscreen.Model,
	appTheme theme.Theme,
) *Model {
	switch {
	case runtime == nil:
		panic("onboarding runtime is required")
	case session == nil:
		panic("onboarding session runtime is required")
	case roleModel == nil:
		panic("onboarding role model is required")
	case tenancyModel == nil:
		panic("onboarding tenancy model is required")
	case integrationsModel == nil:
		panic("onboarding integrations model is required")
	case powersyncModel == nil:
		panic("onboarding powersync model is required")
	}

	model := &Model{
		route:        routeLoading,
		theme:        appTheme,
		runtime:      runtime,
		session:      session,
		role:         roleModel,
		tenancy:      tenancyModel,
		integrations: integrationsModel,
		powersync:    powersyncModel,
	}
	model.router.Register(childRole, model.role)
	model.router.Register(childTenancy, model.tenancy)
	model.router.Register(childIntegrations, model.integrations)
	model.router.Register(childPowerSync, model.powersync)
	return model
}

func (m *Model) Init() tea.Cmd { return m.loadStateCmd() }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case role.SubmittedMsg:
		return m, m.setRoleCmd(msg.Role)
	case tenancyflow.OrganizationSelectedMsg:
		return m, m.selectOrganizationCmd(msg.OrganizationID)
	case tenancyflow.OrganizationCreatedMsg:
		return m, m.createOrganizationCmd(msg.Create)
	case tenancyflow.AccountSelectedMsg:
		return m, m.selectAccountCmd(msg.AccountID)
	case tenancyflow.AccountCreatedMsg:
		return m, m.createAccountCmd(msg.Create)
	case tenancyflow.WorkspaceSelectedMsg:
		return m, m.selectWorkspaceCmd(msg.WorkspaceID)
	case integrationsflow.SetDatadogSiteMsg:
		return m, m.setDatadogSiteCmd(msg.Site)
	case integrationsflow.SubmitDatadogAPIKeyMsg:
		return m, m.submitDatadogAPIKeyCmd(msg.Submission)
	case integrationsflow.SubmitDatadogAppKeyMsg:
		return m, m.submitDatadogAppKeyCmd(msg.Submission)
	case integrationsflow.RefreshRequestedMsg:
		return m, m.loadStateCmd()
	case stateResolvedMsg:
		if msg.err != nil {
			m.loadErr = msg.err
			if m.route == routeLoading {
				m.route = routeError
				m.router.ClearActive()
			}
			return m, nil
		}
		previousRoute := m.route
		m.loadErr = nil
		m.applyState(msg.state)
		cmds := []tea.Cmd{
			m.ensureSessionCmd(msg.state),
		}
		if m.route == routePowerSyncReady {
			if previousRoute != routePowerSyncReady {
				cmds = append(cmds, m.powersync.Init())
			}
			cmds = append(cmds, m.pollStateCmd())
		}
		return m, tea.Batch(cmds...)
	case sessionEnsuredMsg:
		if msg.err != nil {
			m.loadErr = msg.err
			if m.route == routeLoading {
				m.route = routeError
				m.router.ClearActive()
			}
		}
		return m, nil
	case pollStateTickMsg:
		if m.route != routePowerSyncReady {
			return m, nil
		}
		return m, tea.Batch(m.loadStateCmd(), m.pollStateCmd())
	case tea.WindowSizeMsg:
		m.router.SetSizeAll(msg.Width, msg.Height)
		return m, nil
	}

	return m, m.router.Forward(msg)
}

func (m *Model) loadStateCmd() tea.Cmd {
	return func() tea.Msg {
		if m.runtime == nil {
			return stateResolvedMsg{err: fmt.Errorf("onboarding runtime is not configured")}
		}
		state, err := m.runtime.State(context.Background())
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) setRoleCmd(selected preferences.Role) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SetRole(context.Background(), preferences.RoleSelection{Role: selected})
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) selectOrganizationCmd(selected tenancy.OrganizationID) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SelectOrganization(context.Background(), preferences.OrganizationSelection{OrganizationID: selected})
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) createOrganizationCmd(create tenancy.OrganizationCreate) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.CreateOrganization(context.Background(), create)
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) selectAccountCmd(selected tenancy.AccountID) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SelectAccount(context.Background(), preferences.AccountSelection{AccountID: selected})
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) createAccountCmd(create tenancy.AccountCreate) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.CreateAccount(context.Background(), create)
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) selectWorkspaceCmd(selected tenancy.WorkspaceID) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SelectWorkspace(context.Background(), preferences.WorkspaceSelection{WorkspaceID: selected})
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) setDatadogSiteCmd(site integrations.DatadogSite) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SetDatadogSite(context.Background(), site)
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) submitDatadogAPIKeyCmd(submission integrations.DatadogAPIKeySubmission) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SubmitDatadogAPIKey(context.Background(), submission)
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) submitDatadogAppKeyCmd(submission integrations.DatadogAppKeySubmission) tea.Cmd {
	return func() tea.Msg {
		state, err := m.runtime.SubmitDatadogAppKey(context.Background(), submission)
		return stateResolvedMsg{state: state, err: err}
	}
}

func (m *Model) ensureSessionCmd(state onboardingruntime.State) tea.Cmd {
	if state.SelectedOrganization == nil || state.SelectedAccount == nil {
		return nil
	}
	scope := sessionruntime.Scope{
		OrganizationID: state.SelectedOrganization.ID,
		AccountID:      state.SelectedAccount.ID,
	}
	return func() tea.Msg {
		return sessionEnsuredMsg{err: m.session.Ensure(context.Background(), scope)}
	}
}

func (m *Model) pollStateCmd() tea.Cmd {
	return tea.Tick(powersyncPollInterval, func(time.Time) tea.Msg {
		return pollStateTickMsg{}
	})
}

func (m *Model) applyState(state onboardingruntime.State) {
	m.step = state.NextStep
	switch state.NextStep {
	case onboardingruntime.StepRoleSelect:
		m.route = routeRole
		m.router.ActivateOnly(childRole)
	case onboardingruntime.StepPowerSyncReady:
		m.route = routePowerSyncReady
		m.router.ActivateOnly(childPowerSync)
	case onboardingruntime.StepDone:
		m.route = routeDone
		m.router.ClearActive()
	default:
		if m.tenancy.ApplyState(state) {
			m.route = routeTenancy
			m.router.ActivateOnly(childTenancy)
			return
		}
		if m.integrations.ApplyState(state) {
			m.route = routeIntegrations
			m.router.ActivateOnly(childIntegrations)
			return
		}
		m.route = routePlaceholder
		m.router.ClearActive()
	}
}

// ShortHelp returns the active onboarding step key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.router.ShortHelp()
}
