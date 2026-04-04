package busy

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/progressbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns command-bar busy-state rendering.
type Model struct {
	theme    theme.Theme
	width    int
	state    *core.Busy
	active   bool
	progress *progressbar.Model
}

var _ core.Model = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{
		theme:    appTheme,
		progress: progressbar.New(appTheme, 32),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
	if m.progress != nil {
		barWidth := width - 6
		if barWidth < 8 {
			barWidth = 8
		}
		m.progress.SetWidth(barWidth)
	}
}

func (m *Model) ApplyBusy(busy *core.Busy) tea.Cmd {
	m.state = busy
	m.active = busy != nil
	return nil
}

func (m *Model) PreferredHeight(int) int {
	if m.state == nil {
		return 0
	}
	lines := 1
	if strings.TrimSpace(m.state.Status) != "" {
		lines += 2
	}
	if _, ok := m.percent(); ok {
		lines += 2
		if strings.TrimSpace(m.progressDetail()) != "" {
			lines += 2
		}
	}
	if strings.TrimSpace(m.state.Detail) != "" {
		lines += 2
	}
	return lines + 2
}
