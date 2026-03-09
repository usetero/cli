package datadogregion

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

type option struct {
	region integrations.DatadogRegion
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
		scope:   scope,
		theme:   appTheme,
		options: buildOptions(),
		list:    list,
	}
	model.list.SetItems(model.items(), 0)
	return model
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetSize is part of the screen contract.
func (m *Model) SetSize(width, height int) { m.list.SetSize(width, height) }

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
		site := m.options[selectedMsg.Index].region.Site
		m.scope.Info("datadog site highlighted", "site", site)
		return SelectedMsg{Site: site}
	}
}

// View renders the Datadog region selection screen.
func (m *Model) View() tea.View {
	return present.View(
		m.theme,
		present.Stack(
			present.Title("Select your Datadog region"),
			present.Muted("Choose the region where your Datadog account is hosted"),
			present.Raw(""),
			present.Raw(m.renderOptions()),
		),
	)
}

// ShortHelp returns datadog-region key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m *Model) items() []selectlist.Item {
	items := make([]selectlist.Item, 0, len(m.options))
	for i := range m.options {
		items = append(items, selectlist.Item{
			Title: m.options[i].region.DisplayName,
		})
	}
	return items
}

func buildOptions() []option {
	regions := integrations.DatadogRegions()
	out := make([]option, 0, len(regions))
	for _, region := range regions {
		out = append(out, option{region: region})
	}
	return out
}

func (m *Model) renderOptions() string {
	selected := m.list.SelectedIndex()
	lines := make([]string, 0, len(m.options))
	for i := range m.options {
		label := "  " + m.options[i].region.DisplayName
		style := m.theme.Text.Body
		if i == selected {
			label = "> " + m.options[i].region.DisplayName
			style = m.theme.Text.Section
		}
		lines = append(lines, style.Render(label))
	}
	return strings.Join(lines, "\n")
}
