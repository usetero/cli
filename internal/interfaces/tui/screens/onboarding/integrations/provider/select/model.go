package providerselect

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Model owns integration provider selection UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme

	options []integrations.Provider
	list    *selectlist.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the integration provider selection model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	return &Model{scope: scope, theme: appTheme, list: selectlist.New(appTheme)}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the screen contract.
func (m *Model) SetSize(width, height int) { m.list.SetSize(width, height) }

// SetProviders sets selectable providers and resets cursor.
func (m *Model) SetProviders(providers []integrations.Provider) {
	m.options = append([]integrations.Provider(nil), providers...)
	m.list.SetItems(m.items(), 0)
}

// SelectedProvider returns currently highlighted provider.
func (m *Model) SelectedProvider() integrations.Provider {
	index := m.list.SelectedIndex()
	if index < 0 || index >= len(m.options) {
		return ""
	}
	return m.options[index]
}

// Update handles provider selection input.
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
		p := m.options[selectedMsg.Index]
		m.scope.Info("integration provider highlighted", "provider", string(p))
		return SelectedMsg{Provider: p}
	}
}

// View renders the provider selection screen.
func (m *Model) View() tea.View {
	return present.View(m.theme, present.Section("Select your integration provider:", present.Raw(m.list.View().Content)))
}

// ShortHelp returns provider-select key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m *Model) items() []selectlist.Item {
	items := make([]selectlist.Item, 0, len(m.options))
	for i := range m.options {
		items = append(items, selectlist.Item{Title: string(m.options[i])})
	}
	return items
}
