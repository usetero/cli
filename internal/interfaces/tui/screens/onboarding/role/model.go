package role

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

type option struct {
	role        preferences.Role
	label       string
	description string
}

// Model owns role-selection UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme

	options []option
	list    *selectlist.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the onboarding role-selection model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	list := selectlist.New(appTheme)
	model := &Model{
		scope: scope,
		theme: appTheme,
		options: []option{
			{
				role:        preferences.RoleEngineer,
				label:       "Engineer",
				description: "I implement and troubleshoot day-to-day systems.",
			},
			{
				role:        preferences.RolePlatform,
				label:       "Platform",
				description: "I operate shared services and production posture.",
			},
		},
		list: list,
	}
	model.list.SetItems(model.items(), 0)
	return model
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd {
	m.list.SetItems(m.items(), 0)
	return nil
}

// SetSize is part of the screen contract. Role currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// SelectedRole returns the currently highlighted role.
func (m *Model) SelectedRole() preferences.Role {
	index := m.list.SelectedIndex()
	if index < 0 || index >= len(m.options) {
		return ""
	}
	return m.options[index].role
}

// Update handles local role-selection input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.list.Update(msg)
	if listModel, ok := next.(*selectlist.Model); ok {
		m.list = listModel
	}
	if cmd == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		switch selected := cmd().(type) {
		case selectlist.SelectedMsg:
			if selected.Index < 0 || selected.Index >= len(m.options) {
				return nil
			}
			role := m.options[selected.Index].role
			m.scope.Info("role highlighted", "role", role)
			return SubmittedMsg{Role: role}
		default:
			return selected
		}
	}
}

// View renders the role-selection screen.
func (m *Model) View() tea.View {
	return present.View(m.theme, present.Section("Select your role:", present.Raw(m.list.View().Content)))
}

// ShortHelp returns role-screen key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m *Model) items() []selectlist.Item {
	items := make([]selectlist.Item, 0, len(m.options))
	for i := range m.options {
		items = append(items, selectlist.Item{
			Title:    m.options[i].label,
			Subtitle: m.options[i].description,
		})
	}
	return items
}
