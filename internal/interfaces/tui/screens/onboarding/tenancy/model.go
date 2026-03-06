package tenancyflow

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	"github.com/usetero/cli/internal/runtime/onboarding"

	accountcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/create"
	accountselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/select"
	organizationcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/create"
	organizationselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/select"
	workspaceselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/workspace/select"
)

type route int

const (
	routeNone route = iota
	routeOrganizationSelect
	routeOrganizationCreate
	routeAccountSelect
	routeAccountCreate
	routeWorkspaceSelect
)

type childID string

const (
	childOrganizationSelect childID = "organization_select"
	childOrganizationCreate childID = "organization_create"
	childAccountSelect      childID = "account_select"
	childAccountCreate      childID = "account_create"
	childWorkspaceSelect    childID = "workspace_select"
)

type Model struct {
	route  route
	theme  theme.Theme
	router screen.Router[childID]

	organizationSelect *organizationselect.Model
	organizationCreate *organizationcreate.Model
	accountSelect      *accountselect.Model
	accountCreate      *accountcreate.Model
	workspaceSelect    *workspaceselect.Model
}

var _ screen.Model = (*Model)(nil)

func New(
	organizationSelectModel *organizationselect.Model,
	organizationCreateModel *organizationcreate.Model,
	accountSelectModel *accountselect.Model,
	accountCreateModel *accountcreate.Model,
	workspaceSelectModel *workspaceselect.Model,
	appTheme theme.Theme,
) *Model {
	switch {
	case organizationSelectModel == nil:
		panic("tenancy organization select model is required")
	case organizationCreateModel == nil:
		panic("tenancy organization create model is required")
	case accountSelectModel == nil:
		panic("tenancy account select model is required")
	case accountCreateModel == nil:
		panic("tenancy account create model is required")
	case workspaceSelectModel == nil:
		panic("tenancy workspace select model is required")
	}
	model := &Model{
		route:              routeNone,
		theme:              appTheme,
		organizationSelect: organizationSelectModel,
		organizationCreate: organizationCreateModel,
		accountSelect:      accountSelectModel,
		accountCreate:      accountCreateModel,
		workspaceSelect:    workspaceSelectModel,
	}

	model.router.Register(childOrganizationSelect, model.organizationSelect)
	model.router.Register(childOrganizationCreate, model.organizationCreate)
	model.router.Register(childAccountSelect, model.accountSelect)
	model.router.Register(childAccountCreate, model.accountCreate)
	model.router.Register(childWorkspaceSelect, model.workspaceSelect)

	model.router.SetLift(childOrganizationSelect, liftOrganizationSelectCmd)
	model.router.SetLift(childOrganizationCreate, liftOrganizationCreateCmd)
	model.router.SetLift(childAccountSelect, liftAccountSelectCmd)
	model.router.SetLift(childAccountCreate, liftAccountCreateCmd)
	model.router.SetLift(childWorkspaceSelect, liftWorkspaceSelectCmd)

	return model
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) SetSize(width, height int) {
	m.router.SetSizeAll(width, height)
}

func (m *Model) ApplyState(state onboarding.State) bool {
	switch state.NextStep {
	case onboarding.StepOrganizationSelect:
		m.organizationSelect.SetOrganizations(state.Organizations, state.SelectedOrganization)
		m.route = routeOrganizationSelect
		m.router.ActivateOnly(childOrganizationSelect)
		return true
	case onboarding.StepOrganizationCreate:
		m.organizationCreate.Reset()
		m.route = routeOrganizationCreate
		m.router.ActivateOnly(childOrganizationCreate)
		return true
	case onboarding.StepAccountSelect:
		m.accountSelect.SetAccounts(state.Accounts, state.SelectedAccount)
		m.route = routeAccountSelect
		m.router.ActivateOnly(childAccountSelect)
		return true
	case onboarding.StepAccountCreate:
		m.accountCreate.Reset()
		m.route = routeAccountCreate
		m.router.ActivateOnly(childAccountCreate)
		return true
	case onboarding.StepWorkspaceSelect:
		m.workspaceSelect.SetWorkspaces(state.Workspaces, state.SelectedWorkspace)
		m.route = routeWorkspaceSelect
		m.router.ActivateOnly(childWorkspaceSelect)
		return true
	default:
		m.route = routeNone
		m.router.ClearActive()
		return false
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.router.Forward(msg)
}

func (m *Model) View() tea.View {
	switch m.route {
	case routeOrganizationSelect:
		return m.organizationSelect.View()
	case routeOrganizationCreate:
		return m.organizationCreate.View()
	case routeAccountSelect:
		return m.accountSelect.View()
	case routeAccountCreate:
		return m.accountCreate.View()
	case routeWorkspaceSelect:
		return m.workspaceSelect.View()
	default:
		return tea.NewView(lipgloss.JoinVertical(
			lipgloss.Left,
			m.theme.Text.Muted.Render("Tenancy flow is not active."),
		))
	}
}

// ShortHelp returns active tenancy flow key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.router.ShortHelp()
}

func liftOrganizationSelectCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case organizationselect.SelectedMsg:
			return OrganizationSelectedMsg{OrganizationID: typed.OrganizationID}
		default:
			return msg
		}
	}
}

func liftOrganizationCreateCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case organizationcreate.CreatedMsg:
			return OrganizationCreatedMsg{Create: typed.Create}
		default:
			return msg
		}
	}
}

func liftAccountSelectCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case accountselect.SelectedMsg:
			return AccountSelectedMsg{AccountID: typed.AccountID}
		default:
			return msg
		}
	}
}

func liftAccountCreateCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case accountcreate.CreatedMsg:
			return AccountCreatedMsg{Create: typed.Create}
		default:
			return msg
		}
	}
}

func liftWorkspaceSelectCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case workspaceselect.SelectedMsg:
			return WorkspaceSelectedMsg{WorkspaceID: typed.WorkspaceID}
		default:
			return msg
		}
	}
}
