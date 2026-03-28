package onboarding

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	accountcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/account/create"
	accountselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/account/select"
	"github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/auth"
	datadogapikey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/datadog/apikey"
	datadogappkey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/datadog/appkey"
	datadogdiscovery "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/datadog/discovery"
	datadogregion "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/datadog/region"
	"github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/organization/create"
	organizationselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/organization/select"
	powersyncready "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/powersync"
	workspaceselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/workspace/select"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
)

type stateLoadedMsg struct {
	State runtimeonboarding.State
	Err   error
}

type refreshRequestedMsg struct{}

// Model owns the onboarding body flow.
type Model struct {
	core.Router

	scope       logging.Scope
	identity    *identity.Service
	workflow    *runtimeonboarding.Workflow
	auth        *auth.Model
	orgSelect   *organizationselect.Model
	orgCreate   *create.Model
	acctSelect  *accountselect.Model
	acctCreate  *accountcreate.Model
	wsSelect    *workspaceselect.Model
	ddDiscovery *datadogdiscovery.Model
	ddRegion    *datadogregion.Model
	ddAPIKey    *datadogapikey.Model
	ddAppKey    *datadogappkey.Model
	psReady     *powersyncready.Model
	loading     bool
	busy        *core.Busy
	state       runtimeonboarding.State
	account     accountruntime.Status
	placeholder *core.Input
}

var _ core.Model = (*Model)(nil)
var _ core.BusyProvider = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)
var _ core.HelpProvider = (*Model)(nil)

func New(scope logging.Scope, identityService *identity.Service, workflow *runtimeonboarding.Workflow, appTheme theme.Theme) *Model {
	if workflow == nil {
		panic("onboarding model requires workflow")
	}

	authModel := auth.New(scope.Child("auth"), identityService, appTheme)
	orgSelect := organizationselect.New(appTheme)
	orgCreate := create.New(appTheme)
	acctSelect := accountselect.New(appTheme)
	acctCreate := accountcreate.New(appTheme)
	wsSelect := workspaceselect.New(appTheme)
	ddDiscovery := datadogdiscovery.New(appTheme)
	ddRegion := datadogregion.New(appTheme)
	ddAPIKey := datadogapikey.New(appTheme)
	ddAppKey := datadogappkey.New(appTheme)
	psReady := powersyncready.New(appTheme)

	return &Model{
		Router:      core.Router{},
		scope:       scope,
		identity:    identityService,
		workflow:    workflow,
		auth:        authModel,
		orgSelect:   orgSelect,
		orgCreate:   orgCreate,
		acctSelect:  acctSelect,
		acctCreate:  acctCreate,
		wsSelect:    wsSelect,
		ddDiscovery: ddDiscovery,
		ddRegion:    ddRegion,
		ddAPIKey:    ddAPIKey,
		ddAppKey:    ddAppKey,
		psReady:     psReady,
		placeholder: &core.Input{
			Label: "Next onboarding steps are not built yet.",
		},
	}
}

func (m *Model) Init() tea.Cmd {
	m.showAuth()
	if m.identity != nil && m.identity.IsAuthenticated() {
		m.loading = true
		return m.loadState()
	}
	return m.Router.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case stateLoadedMsg:
		m.loading = false
		m.busy = nil
		if typed.Err != nil {
			m.scope.Error("load onboarding state", "error", typed.Err)
			return m, nil
		}
		return m, m.applyState(typed.State)
	case refreshRequestedMsg:
		if !m.shouldRefresh() {
			return m, nil
		}
		return m, m.loadState()
	case events.AccountRuntimeUpdatedMsg:
		m.account = typed.Status
		m.psReady.SetStatus(typed.Status)
		if m.state.NextStep == runtimeonboarding.StepDone || m.Active() == m.psReady {
			return m, m.routeState(m.state)
		}
		return m, nil
	case organizationselect.SelectedMsg:
		m.busy = &core.Busy{Label: "Selecting Organization"}
		return m, m.selectOrganization(typed.OrganizationID)
	case create.CreatedMsg:
		m.busy = &core.Busy{Label: "Creating Organization"}
		return m, m.createOrganization(typed.Name)
	case accountselect.SelectedMsg:
		m.busy = &core.Busy{Label: "Selecting Account"}
		return m, m.selectAccount(typed.AccountID)
	case accountcreate.CreatedMsg:
		m.busy = &core.Busy{Label: "Creating Account"}
		return m, m.createAccount(typed.Name)
	case workspaceselect.SelectedMsg:
		m.busy = &core.Busy{Label: "Selecting Workspace"}
		return m, m.selectWorkspace(typed.WorkspaceID)
	case datadogregion.SelectedMsg:
		m.busy = &core.Busy{Label: "Saving Datadog Region"}
		return m, m.setDatadogSite(typed.Site)
	case datadogapikey.SubmittedMsg:
		m.busy = &core.Busy{Label: "Validating Datadog API Key"}
		return m, m.submitDatadogAPIKey(typed.APIKey)
	case datadogappkey.SubmittedMsg:
		m.busy = &core.Busy{Label: "Saving Datadog App Key"}
		return m, m.submitDatadogAppKey(typed.AppKey)
	}

	switch m.Active() {
	case m.auth:
		cmd := m.Router.Update(msg)
		if m.auth.Authenticated() && !m.loading {
			m.loading = true
			return m, tea.Batch(cmd, m.loadState())
		}
		return m, cmd
	case m.orgSelect, m.orgCreate, m.acctSelect, m.acctCreate, m.wsSelect, m.ddDiscovery, m.ddRegion, m.ddAPIKey, m.ddAppKey, m.psReady:
		return m, m.Router.Update(msg)
	default:
		return m, nil
	}
}

func (m *Model) loadState() tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.State(context.Background())
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) applyState(state runtimeonboarding.State) tea.Cmd {
	m.state = state

	var cmds []tea.Cmd
	if state.SelectedOrganization != nil {
		org := *state.SelectedOrganization
		cmds = append(cmds, func() tea.Msg { return events.OrganizationSelectedMsg{Organization: org} })
	}
	if state.SelectedOrganization != nil && state.SelectedAccount != nil {
		account := *state.SelectedAccount
		scope := accountruntime.Scope{
			Organization: *state.SelectedOrganization,
			Account:      account,
		}
		if state.SelectedWorkspace != nil {
			scope.Workspace = *state.SelectedWorkspace
		}
		cmds = append(cmds, func() tea.Msg { return events.AccountSelectedMsg{Scope: scope} })
	}

	cmds = append(cmds, m.routeState(state))
	return tea.Batch(cmds...)
}

func (m *Model) routeState(state runtimeonboarding.State) tea.Cmd {
	var cmds []tea.Cmd

	switch state.NextStep {
	case runtimeonboarding.StepOrganizationSelect:
		m.orgSelect.SetOrganizations(state.Organizations)
		m.showOrganizationSelect()
	case runtimeonboarding.StepOrganizationCreate:
		m.showOrganizationCreate()
	case runtimeonboarding.StepAccountSelect:
		m.acctSelect.SetAccounts(state.Accounts)
		m.showAccountSelect()
	case runtimeonboarding.StepAccountCreate:
		m.showAccountCreate()
	case runtimeonboarding.StepWorkspaceSelect:
		m.wsSelect.SetWorkspaces(state.Workspaces)
		m.showWorkspaceSelect()
	case runtimeonboarding.StepDatadogDiscovery:
		m.ddDiscovery.SetStatus(state.DatadogStatus)
		m.showDatadogDiscovery()
		cmds = append(cmds, m.refreshAfter(2*time.Second))
	case runtimeonboarding.StepDatadogRegion:
		m.showDatadogRegion()
	case runtimeonboarding.StepDatadogAPIKey:
		m.ddAPIKey.SetSite(state.DatadogDraft.Site)
		m.showDatadogAPIKey()
	case runtimeonboarding.StepDatadogAppKey:
		m.ddAppKey.SetSite(state.DatadogDraft.Site)
		m.showDatadogAppKey()
	case runtimeonboarding.StepPowerSyncReady:
		m.psReady.SetStatus(m.account)
		m.showPowerSyncReady()
	case runtimeonboarding.StepDone:
		if m.shouldWaitForInitialSync(state) {
			m.psReady.SetStatus(m.account)
			m.showPowerSyncReady()
		} else {
			m.clear()
		}
	default:
		m.clear()
	}

	return tea.Batch(cmds...)
}

func (m *Model) selectOrganization(id tenancy.OrganizationID) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.SelectOrganization(context.Background(), organizationselect.Selection(string(id)))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) createOrganization(name string) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.CreateOrganization(context.Background(), create.Submission(name))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) selectAccount(id tenancy.AccountID) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.SelectAccount(context.Background(), accountselect.Selection(string(id)))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.CreateAccount(context.Background(), accountcreate.Submission(name))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) selectWorkspace(id tenancy.WorkspaceID) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.SelectWorkspace(context.Background(), workspaceselect.Selection(string(id)))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) setDatadogSite(site integrations.DatadogSite) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.SetDatadogSite(context.Background(), site)
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) submitDatadogAPIKey(value string) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.SubmitDatadogAPIKey(context.Background(), datadogapikey.Submission(value))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) submitDatadogAppKey(value string) tea.Cmd {
	return func() tea.Msg {
		state, err := m.workflow.SubmitDatadogAppKey(context.Background(), datadogappkey.Submission(value))
		return stateLoadedMsg{State: state, Err: err}
	}
}

func (m *Model) showAuth() {
	m.Router.SetActive(m.auth)
}

func (m *Model) showOrganizationSelect() {
	m.Router.SetActive(m.orgSelect)
}

func (m *Model) showOrganizationCreate() {
	m.Router.SetActive(m.orgCreate)
}

func (m *Model) showAccountSelect() {
	m.Router.SetActive(m.acctSelect)
}

func (m *Model) showAccountCreate() {
	m.Router.SetActive(m.acctCreate)
}

func (m *Model) showWorkspaceSelect() {
	m.Router.SetActive(m.wsSelect)
}

func (m *Model) showDatadogDiscovery() {
	m.Router.SetActive(m.ddDiscovery)
}

func (m *Model) showDatadogRegion() {
	m.Router.SetActive(m.ddRegion)
}

func (m *Model) showDatadogAPIKey() {
	m.Router.SetActive(m.ddAPIKey)
}

func (m *Model) showDatadogAppKey() {
	m.Router.SetActive(m.ddAppKey)
}

func (m *Model) showPowerSyncReady() {
	m.Router.SetActive(m.psReady)
}

func (m *Model) clear() {
	m.Router.SetActive(nil)
}

func (m *Model) shouldRefresh() bool {
	switch m.Active() {
	case m.ddDiscovery:
		return true
	default:
		return false
	}
}

func (m *Model) refreshAfter(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return refreshRequestedMsg{}
	})
}

func (m *Model) shouldWaitForInitialSync(state runtimeonboarding.State) bool {
	if state.SelectedAccount == nil {
		return false
	}
	return !m.account.HasCompletedInitialSync
}
