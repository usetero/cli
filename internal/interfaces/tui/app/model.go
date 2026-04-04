package app

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	accountmodel "github.com/usetero/cli/internal/interfaces/tui/app/account"
	chromedivider "github.com/usetero/cli/internal/interfaces/tui/components/chrome/divider"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

var quitBinding = key.NewBinding(
	key.WithKeys("ctrl+c"),
	key.WithHelp("ctrl+c", "quit"),
)

const windowTitle = "Tero"

// Model renders the TUI shell.
type Model struct {
	scope            logging.Scope
	theme            theme.Theme
	width            int
	height           int
	account          *accountmodel.Model
	surface          surface
	titleSpinnerTick bool
	titleSpinnerStep int
}

func New(scope logging.Scope, runtimeFactory accountmodel.RuntimeFactory, body core.Screen, env string, appTheme theme.Theme) *Model {
	if runtimeFactory == nil {
		panic("app runtime factory is required")
	}
	if body == nil {
		panic("app body model is required")
	}

	m := &Model{
		scope:   scope,
		theme:   appTheme,
		account: accountmodel.New(scope, runtimeFactory),
		surface: newSurface(
			body,
			statusbar.New(env, appTheme),
			chromedivider.New(appTheme),
			commandbar.New(appTheme),
			helpbar.New(appTheme),
		),
	}
	m.applyLayout()
	return m
}

func (m *Model) Init() tea.Cmd {
	m.scope.Info("tui initialized")
	cmds := []tea.Cmd{m.surface.Init()}
	if cmd := m.startWindowTitleSpinner(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.SetSize(size.Width, size.Height)
	}

	switch typed := msg.(type) {
	case core.CommandSelectedMsg:
		switch typed.ID {
		case core.CommandQuit:
			m.scope.Info("quit requested", "mode", "command")
			_ = m.account.Close(ctx)
			return m, tea.Quit
		case core.CommandOpenSpikes:
			m.surface.shell.commandbar.SetLocalNotice(&core.Notice{
				Level:   core.NoticeInfo,
				Message: "Spikes view is not wired yet.",
			})
			m.applyLayout()
			return m, nil
		case core.CommandOpenWaste:
			m.surface.shell.commandbar.SetLocalNotice(&core.Notice{
				Level:   core.NoticeInfo,
				Message: "Waste view is not wired yet.",
			})
			m.applyLayout()
			return m, nil
		case core.CommandOpenServices:
			m.surface.shell.commandbar.SetLocalNotice(&core.Notice{
				Level:   core.NoticeInfo,
				Message: "Services view is not wired yet.",
			})
			m.applyLayout()
			return m, nil
		}
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, quitBinding) {
		m.scope.Info("quit requested", "mode", "immediate")
		_ = m.account.Close(ctx)
		return m, tea.Quit
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if shellCmd, consumed := m.surface.HandleKey(keyMsg); consumed {
			m.applyLayout()

			var titleCmd tea.Cmd
			if m.surface.body.Busy() != nil {
				titleCmd = m.startWindowTitleSpinner()
			} else {
				m.stopWindowTitleSpinner()
			}

			return m, tea.Batch(shellCmd, titleCmd)
		}
	}

	if cmd := m.updateWindowTitleSpinner(msg); cmd != nil || m.surface.body.Busy() == nil {
		if _, ok := msg.(windowTitleTickMsg); ok {
			return m, cmd
		}
	}

	_, accountCmd := m.account.Update(msg)
	surfaceCmd := m.surface.Update(msg)
	m.applyLayout()

	var titleCmd tea.Cmd
	if m.surface.body.Busy() != nil {
		titleCmd = m.startWindowTitleSpinner()
	} else {
		m.stopWindowTitleSpinner()
	}

	return m, tea.Batch(accountCmd, surfaceCmd, titleCmd)
}
