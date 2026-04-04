package datadogregion

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns the Datadog-region selection page state.
type Model struct {
	options []core.Option
}

var _ core.Model = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)

func New(theme.Theme) *Model {
	m := &Model{}
	m.setRegions(integrations.DatadogRegions())
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case selectlist.SelectedMsg:
		return m, func() tea.Msg { return SelectedMsg{Site: integrations.DatadogSite(typed.Option.ID)} }
	default:
		return m, nil
	}
}

func (m *Model) View() tea.View { return tea.NewView("") }

func (m *Model) SetSize(width, height int) {}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Kind:    core.InputSelect,
		Title:   "Choose your Datadog region.",
		Options: append([]core.Option(nil), m.options...),
	}
}

func (m *Model) setRegions(regions []integrations.DatadogRegion) {
	options := make([]core.Option, 0, len(regions))
	for _, region := range regions {
		options = append(options, core.Option{
			ID:       string(region.Site),
			Label:    region.DisplayName,
			Subtitle: region.Description,
		})
	}
	m.options = options
}
