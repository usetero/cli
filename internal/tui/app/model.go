package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	appchat "github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/tools"
	"github.com/usetero/cli/internal/tui/app/tools/endjourney"
	"github.com/usetero/cli/internal/tui/app/tools/query"
	"github.com/usetero/cli/internal/tui/app/tools/startjourney"
)

// Model is the app shell.
// It's a thin router that delegates to the current page.
type Model struct {
	ctx    context.Context
	theme  *styles.Theme
	logger log.Logger

	// Current page (for now, always chat)
	chat appchat.Model
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

	appTools := tools.Tools{
		StartJourney: &startjourney.Tool{},
		EndJourney:   &endjourney.Tool{},
		Query:        &query.Tool{DB: db},
	}

	return Model{
		ctx:    ctx,
		theme:  theme,
		logger: logger,
		chat:   appchat.New(ctx, theme, db, client, account.ID, workspace.ID, appTools, logger),
	}
}

// Init initializes the app.
func (m Model) Init() tea.Cmd {
	return m.chat.Init()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chat, cmd = m.chat.Update(msg)
	return m, cmd
}

// View renders the app.
func (m Model) View() string {
	return m.chat.View()
}
