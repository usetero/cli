package app

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	accountmodel "github.com/usetero/cli/internal/interfaces/tui/app/account"
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
	scope    logging.Scope
	theme    theme.Theme
	width    int
	height   int
	account  *accountmodel.Model
	body     core.Screen
	children children
}

func New(scope logging.Scope, runtimeFactory accountmodel.RuntimeFactory, body core.Screen, env string, appTheme theme.Theme) *Model {
	if runtimeFactory == nil {
		panic("app runtime factory is required")
	}
	if body == nil {
		panic("app body model is required")
	}

	status := statusbar.New(env, appTheme)
	command := commandbar.New(appTheme)
	help := helpbar.New(appTheme)
	m := &Model{
		scope:    scope,
		theme:    appTheme,
		account:  accountmodel.New(scope, runtimeFactory),
		body:     body,
		children: newChildren(status, command, help),
	}
	m.children.refreshFooter(m.body)
	m.applyLayout()
	return m
}

func (m *Model) Init() tea.Cmd {
	m.scope.Info("tui initialized")
	m.children.refreshFooter(m.body)
	return tea.Batch(m.children.Init(), m.body.Init())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.SetSize(size.Width, size.Height)
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, quitBinding) {
		m.scope.Info("quit requested", "mode", "immediate")
		_ = m.account.Close(ctx)
		return m, tea.Quit
	}

	_, accountCmd := m.account.Update(msg)
	nextBody, bodyCmd := m.body.Update(msg)
	if typed, ok := nextBody.(core.Screen); ok && typed != nil {
		m.body = typed
	}
	shellCmd := m.children.Update(msg)
	m.children.refreshFooter(m.body)
	m.applyLayout()
	return m, tea.Batch(accountCmd, bodyCmd, shellCmd)
}
