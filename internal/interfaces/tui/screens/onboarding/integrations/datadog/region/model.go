package datadogregion

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

type option struct {
	site integrations.DatadogSite
}

// Model owns Datadog region selection UI state.
type Model struct {
	scope   logging.Scope
	theme   theme.Theme
	options []option
	list    *selectlist.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the Datadog region selection model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	list := selectlist.New(appTheme)
	model := &Model{
		scope: scope,
		theme: appTheme,
		options: []option{
			{site: integrations.DatadogSiteUS1},
			{site: integrations.DatadogSiteUS3},
			{site: integrations.DatadogSiteUS5},
			{site: integrations.DatadogSiteEU1},
			{site: integrations.DatadogSiteUS1Fed},
			{site: integrations.DatadogSiteAP1},
			{site: integrations.DatadogSiteAP2},
		},
		list: list,
	}
	model.list.SetItems(model.items(), 0)
	return model
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetSize is part of the screen contract. Region select currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// Update handles local region-selection input.
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
		site := m.options[selectedMsg.Index].site
		m.scope.Info("datadog site highlighted", "site", site)
		return SelectedMsg{Site: site}
	}
}

// View renders the Datadog region selection screen.
func (m *Model) View() tea.View {
	lines := []string{
		m.theme.Text.Section.Render("Select your Datadog site:"),
		"",
		m.theme.Text.Body.Render(m.list.View().Content),
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// ShortHelp returns datadog-region key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m *Model) items() []selectlist.Item {
	items := make([]selectlist.Item, 0, len(m.options))
	for i := range m.options {
		items = append(items, selectlist.Item{Title: string(m.options[i].site)})
	}
	return items
}
