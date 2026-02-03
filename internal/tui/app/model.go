package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	appchat "github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/messages"
	"github.com/usetero/cli/internal/tui/app/tools"
)

// Model is the app shell.
// It owns the command bar and routes messages to the focused page.
type Model struct {
	ctx    context.Context
	theme  *styles.Theme
	db     sqlite.DB
	logger log.Logger

	// Components
	commandBar CommandBar
	chat       appchat.Model

	// TODO: focused page (for now, always chat)

	width  int
	height int
}

// New creates a new app model.
func New(ctx context.Context, theme *styles.Theme, db sqlite.DB, client chat.Client, account domain.Account, workspace domain.Workspace, logger log.Logger) Model {
	if ctx == nil {
		panic("app.New: ctx is nil")
	}
	if theme == nil {
		panic("app.New: theme is nil")
	}
	if db == nil {
		panic("app.New: db is nil")
	}
	if client == nil {
		panic("app.New: client is nil")
	}
	if logger == nil {
		panic("app.New: logger is nil")
	}

	turn := appchat.NewTurn(client, logger)

	return Model{
		ctx:        ctx,
		theme:      theme,
		db:         db,
		logger:     logger,
		commandBar: NewCommandBar(theme),
		chat:       appchat.New(ctx, theme, db, turn, account.ID, workspace.ID, tools.All(), logger),
	}
}

// Init initializes the app.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.commandBar.Init(),
		m.chat.Init(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updateLayout()
		return m, nil

	case messages.SubmitMsg:
		// Propagate to focused page (currently always chat)
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		return m, cmd
	}

	// Update command bar
	var cmd tea.Cmd
	m.commandBar, cmd = m.commandBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update chat (for its internal messages like streaming events)
	m.chat, cmd = m.chat.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the app.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	chatView := m.chat.View()
	commandBarView := m.commandBar.View()

	return lipgloss.JoinVertical(lipgloss.Left, chatView, commandBarView)
}

// updateLayout updates component sizes.
func (m Model) updateLayout() Model {
	commandBarHeight := m.commandBar.Height()
	chatHeight := m.height - commandBarHeight

	m.commandBar = m.commandBar.SetWidth(m.width)
	m.chat = m.chat.SetSize(m.width, chatHeight)

	return m
}
