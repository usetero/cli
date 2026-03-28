package statusbar

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

// Model renders the global status bar header.
type Model struct {
	status               accountruntime.Status
	selectedOrganization tenancy.Organization
	accountSelected      bool
	env                  string
	theme                theme.Theme
	width                int
}

var _ core.Model = (*Model)(nil)

// New constructs a status bar model.
func New(env string, appTheme theme.Theme) *Model {
	return &Model{
		env:   env,
		theme: appTheme,
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case events.OrganizationSelectedMsg:
		m.selectedOrganization = typed.Organization
	case events.AccountSelectedMsg:
		m.accountSelected = true
		if typed.Scope.Organization.ID != "" {
			m.selectedOrganization = typed.Scope.Organization
		}
	case events.AccountRuntimeUpdatedMsg:
		m.status = typed.Status
	}
	return m, nil
}
