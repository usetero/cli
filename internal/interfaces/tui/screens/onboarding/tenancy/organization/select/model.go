package organizationselect

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Model owns organization-selection UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme

	options []tenancy.Organization
	list    *selectlist.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the onboarding organization-selection model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	return &Model{scope: scope, theme: appTheme, list: selectlist.New(appTheme)}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetSize is part of the screen contract. Organization currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// SetOrganizations replaces the selectable organizations and resets cursor if needed.
func (m *Model) SetOrganizations(orgs []tenancy.Organization, selected *tenancy.Organization) {
	m.options = append([]tenancy.Organization(nil), orgs...)
	selectedIndex := 0
	if selected == nil {
		m.list.SetItems(m.items(), selectedIndex)
		return
	}
	for i := range m.options {
		if m.options[i].ID == selected.ID {
			selectedIndex = i
			break
		}
	}
	m.list.SetItems(m.items(), selectedIndex)
}

// SelectedOrganizationID returns the currently highlighted organization id.
func (m *Model) SelectedOrganizationID() tenancy.OrganizationID {
	index := m.list.SelectedIndex()
	if index < 0 || index >= len(m.options) {
		return ""
	}
	return m.options[index].ID
}

// Update handles local organization-selection input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.list.Update(msg)
	if model, ok := next.(*selectlist.Model); ok {
		m.list = model
	}
	if cmd == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		selectedMsg, ok := cmd().(selectlist.SelectedMsg)
		if !ok || selectedMsg.Index < 0 || selectedMsg.Index >= len(m.options) {
			return nil
		}
		selectedID := m.options[selectedMsg.Index].ID
		m.scope.Info("organization highlighted", "organization_id", selectedID)
		return SelectedMsg{OrganizationID: selectedID}
	}
}

// View renders the organization-selection screen.
func (m *Model) View() tea.View {
	lines := []string{m.theme.Text.Section.Render("Select your organization:")}

	if len(m.options) == 0 {
		lines = append(lines, "", m.theme.Text.Muted.Render("No organizations available."))
		return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	lines = append(lines, "", m.theme.Text.Body.Render(m.list.View().Content))
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// ShortHelp returns organization-select key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m *Model) items() []selectlist.Item {
	items := make([]selectlist.Item, 0, len(m.options))
	for i := range m.options {
		items = append(items, selectlist.Item{
			Title:    m.options[i].Name,
			Subtitle: "ID: " + string(m.options[i].ID),
		})
	}
	return items
}
