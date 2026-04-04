package commandbar

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/busy"
	errorview "github.com/usetero/cli/internal/interfaces/tui/components/commandbar/error"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/input"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/palette"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type actionSurface struct {
	router     core.Router
	busy       *busy.Model
	err        *errorview.Model
	input      *input.Model
	palette    *palette.Model
	selectlist *selectlist.Model
}

func newActionSurface(appTheme theme.Theme) *actionSurface {
	return &actionSurface{
		busy:       busy.New(appTheme),
		err:        errorview.New(appTheme),
		input:      input.New(appTheme),
		palette:    palette.New(appTheme),
		selectlist: selectlist.New(appTheme),
	}
}

func (s *actionSurface) Init() tea.Cmd {
	return tea.Batch(
		s.busy.Init(),
		s.err.Init(),
		s.input.Init(),
		s.palette.Init(),
		s.selectlist.Init(),
	)
}

func (s *actionSurface) Update(msg tea.Msg) tea.Cmd {
	return s.router.Update(msg)
}

func (s *actionSurface) Active() core.Model {
	return s.router.Active()
}

func (s *actionSurface) SetActive(model core.Model) {
	s.router.SetActive(model)
}

func (s *actionSurface) View() tea.View {
	return s.router.View()
}

func (s *actionSurface) SetSize(width, height int) {
	s.busy.SetSize(width, height)
	s.err.SetSize(width, height)
	s.input.SetSize(width, height)
	s.palette.SetSize(width, height)
	s.selectlist.SetSize(width, height)
	s.router.SetSize(width, height)
}

func (s *actionSurface) ShortHelp() []key.Binding {
	return s.router.ShortHelp()
}

func (s *actionSurface) CapturingInput() bool {
	type captureProvider interface {
		CapturingInput() bool
	}
	if provider, ok := s.router.Active().(captureProvider); ok {
		return provider.CapturingInput()
	}
	return false
}

func (s *actionSurface) HandleKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	type keyConsumer interface {
		ConsumesKey(tea.KeyPressMsg) bool
	}
	active := s.router.Active()
	if active == nil {
		return nil, false
	}
	if consumer, ok := active.(keyConsumer); ok && !consumer.ConsumesKey(msg) {
		return nil, false
	}
	return s.router.Update(msg), true
}

