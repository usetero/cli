package error

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns command-bar error-state rendering.
type Model struct {
	theme  theme.Theme
	width  int
	state  *core.Error
}

var _ core.Model = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
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
}

func (m *Model) ApplyError(state *core.Error) {
	m.state = state
}

func (m *Model) PreferredHeight(int) int {
	if m.state == nil {
		return 0
	}
	lines := 1
	if strings.TrimSpace(m.state.Message) != "" {
		lines += 2
	}
	if strings.TrimSpace(m.state.Detail) != "" {
		lines += 2
	}
	if strings.TrimSpace(m.state.Action) != "" {
		lines += 2
	}
	return lines + 2
}
