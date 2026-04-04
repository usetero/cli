package understanding

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	readmodel "github.com/usetero/cli/internal/readmodels/understanding"
)

// Model owns the Understanding screen.
type Model struct {
	theme  theme.Theme
	width  int
	height int
}

var _ core.Screen = (*Model)(nil)

func New(appTheme theme.Theme, reader readmodel.Reader) *Model {
	_ = reader
	return &Model{
		theme: appTheme,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(typed.Width, typed.Height)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Page() core.Page {
	return core.Page{Title: "Understanding"}
}

func (m *Model) Commands() []core.Command {
	return []core.Command{
		{
			ID:          core.CommandOpenSpikes,
			Title:       "Spikes",
			Description: "Review acute telemetry problems that need attention now",
		},
		{
			ID:          core.CommandOpenWaste,
			Title:       "Waste",
			Description: "Review ongoing telemetry waste and justification pressure",
		},
		{
			ID:          core.CommandOpenServices,
			Title:       "Services",
			Description: "Browse services and drill into local operating context",
		},
	}
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "ask")),
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "group")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "color")),
		key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "metric")),
	}
}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Kind:        core.InputMultiline,
		Title:       "Hi Ben, I'm Tero.",
		Detail:      "I've mapped your telemetry estate and found something acute: checkout-api is logging in the hot path, driving about 2.3m events per hour. I can break it down, regroup the system, or help you fix it now.",
		Placeholder: "Ask what's noisy, wasteful, or changing...",
	}
}

func (m *Model) Busy() *core.Busy   { return nil }
func (m *Model) Error() *core.Error { return nil }
func (m *Model) Notice() *core.Notice { return nil }
