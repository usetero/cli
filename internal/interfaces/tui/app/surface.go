package app

import (
	tea "charm.land/bubbletea/v2"
	chromedivider "github.com/usetero/cli/internal/interfaces/tui/components/chrome/divider"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

type surface struct {
	body  core.Screen
	shell shell
}

func newSurface(body core.Screen, status *statusbar.Model, divider *chromedivider.Model, command *commandbar.Model, help *helpbar.Model) surface {
	s := surface{
		body:  body,
		shell: newShell(status, divider, command, help),
	}
	s.shell.UpdateState(s.body)
	return s
}

func (s *surface) Init() tea.Cmd {
	return tea.Batch(s.shell.Init(), s.body.Init())
}

func (s *surface) Update(msg tea.Msg) tea.Cmd {
	nextBody, bodyCmd := s.body.Update(msg)
	if typed, ok := nextBody.(core.Screen); ok && typed != nil {
		s.body = typed
	}
	s.shell.UpdateState(s.body)
	shellCmd := s.shell.Update(msg)
	return tea.Batch(bodyCmd, shellCmd)
}

func (s *surface) HandleKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	cmd, consumed := s.shell.commandbar.HandleKey(msg)
	if consumed {
		s.shell.UpdateState(s.body)
	}
	return cmd, consumed
}
