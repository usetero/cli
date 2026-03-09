package spinner

import (
	bubblespinner "charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Model wraps the Bubble spinner with shared TUI styling.
type Model struct {
	spinner bubblespinner.Model
}

// New constructs a spinner model.
func New(appTheme theme.Theme) *Model {
	sp := bubblespinner.New()
	sp.Spinner = bubblespinner.Dot
	sp.Style = appTheme.Text.Body.Foreground(appTheme.AccentAlt)
	return &Model{spinner: sp}
}

// Init starts spinner animation.
func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update advances spinner animation.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	tick, ok := msg.(bubblespinner.TickMsg)
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(tick)
	return m, cmd
}

// View renders the current spinner frame.
func (m *Model) View() tea.View { return tea.NewView(m.spinner.View()) }
