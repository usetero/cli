package app

import (
	"context"
	"sort"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/commandbar"
	"github.com/usetero/cli/internal/tui/components/header"
	"github.com/usetero/cli/internal/tui/components/sidebar"
)

// tokenProvider provides access tokens for API authentication.
type tokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

const (
	// Width threshold for switching between sidebar and header mode
	CompactModeWidth = 120

	// Sidebar width when in wide mode
	SidebarWidth = 32
)

// App is the main application orchestrator.
// It renders pages with appropriate chrome (sidebar or header) based on
// window size, and manages the command bar and popover stack.
type App struct {
	// Theme for styling
	theme *styles.Theme

	// Base layer - always chat
	chat page.Page

	// Popover stack - pages layered on top of chat
	popoverStack []page.Page

	// Chrome components
	sidebar    *sidebar.Sidebar
	header     *header.Header
	commandbar *commandbar.CommandBar

	// Dependencies
	logger        log.Logger
	tokenProvider tokenProvider

	// Data layer
	powersyncConfig *powersync.Config
	db              sqlite.Database
	sync            *powersync.Sync

	// Identity
	org     api.Organization
	account api.Account

	// Global key bindings (for footer display, intercepted by tui)
	globalBindings []key.Binding

	// Dimensions
	width  int
	height int

	// State
	compact bool // true when width < CompactModeWidth
}

// New creates a new app starting with the chat page
func New(theme *styles.Theme, org api.Organization, account api.Account, tokenProvider tokenProvider, powersyncConfig *powersync.Config, logger log.Logger, globalBindings []key.Binding) *App {
	return &App{
		theme:           theme,
		chat:            chat.New(theme, org.ID, account.ID, logger),
		sidebar:         sidebar.New(theme, logger),
		header:          header.New(theme, logger),
		commandbar:      commandbar.New(theme, logger),
		logger:          logger,
		tokenProvider:   tokenProvider,
		powersyncConfig: powersyncConfig,
		sync:            powersync.NewSync(powersyncConfig),
		org:             org,
		account:         account,
		globalBindings:  globalBindings,
	}
}

// syncStartedMsg is sent when PowerSync has started.
type syncStartedMsg struct {
	db  sqlite.Database
	err error
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.chat.Init(),
		a.commandbar.Init(),
		a.startSync(),
	)
}

// startSync opens the database and starts PowerSync in the background.
func (a *App) startSync() tea.Cmd {
	return func() tea.Msg {
		// Get database path
		dbPath, err := a.powersyncConfig.DatabasePath(a.account.ID)
		if err != nil {
			return syncStartedMsg{err: err}
		}

		// Open database
		db, err := sqlite.Open(dbPath)
		if err != nil {
			return syncStartedMsg{err: err}
		}

		// Get auth token
		ctx := context.Background()
		token, err := a.tokenProvider.GetAccessToken(ctx)
		if err != nil {
			db.Close()
			return syncStartedMsg{err: err}
		}

		// Start PowerSync
		if err := a.sync.Start(ctx, db, a.account.ID, token); err != nil {
			db.Close()
			return syncStartedMsg{err: err}
		}

		return syncStartedMsg{db: db}
	}
}

// activePage returns the topmost active page (popover or chat)
func (a *App) activePage() page.Page {
	if len(a.popoverStack) > 0 {
		return a.popoverStack[len(a.popoverStack)-1]
	}
	return a.chat
}

// Update handles messages and routes to the appropriate layer
func (a *App) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case syncStartedMsg:
		if msg.err != nil {
			a.logger.Error("failed to start sync", "error", msg.err)
			// TODO: Show error in UI
			return nil
		}
		a.db = msg.db
		a.logger.Info("sync started", "status", a.sync.Status())
		// Notify chat that database is ready
		return func() tea.Msg {
			return chat.DatabaseReadyMsg{DB: msg.db}
		}

	case tea.KeyPressMsg:
		// Check for popover dismiss (Esc)
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) && len(a.popoverStack) > 0 {
			a.popoverStack = a.popoverStack[:len(a.popoverStack)-1]
			a.updateChrome()
			return nil
		}
	}

	// Update command bar
	cmd := a.commandbar.Update(msg)
	cmds = append(cmds, cmd)

	// Route to active page
	cmd = a.activePage().Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// updateChrome updates sidebar/header with current page's metadata
func (a *App) updateChrome() {
	p := a.activePage()

	// Sort metadata by priority
	meta := p.Metadata()
	sort.Slice(meta, func(i, j int) bool {
		return meta[i].Priority < meta[j].Priority
	})

	// Update sidebar with all metadata
	a.sidebar.SetTitle(p.Title())
	a.sidebar.SetMetadata(meta)
	a.sidebar.SetOrgName(a.org.Name)

	// Update header with limited metadata (top priority items)
	a.header.SetTitle(p.Title())
	a.header.SetMetadata(meta)
	a.header.SetOrgName(a.org.Name)

	// Update command bar with page capabilities
	a.commandbar.SetAcceptsNaturalLanguage(p.AcceptsNaturalLanguage())
	a.commandbar.SetCommands(p.Commands())

	// Combine page keybindings with global bindings for footer display
	allBindings := append(p.KeyBindings(), a.globalBindings...)
	a.commandbar.SetKeyBindings(allBindings)
}

// View renders the app
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return ""
	}

	colors := a.theme.Colors
	a.updateChrome()

	// Get command bar view and height
	commandbarView := a.commandbar.View()
	commandbarHeight := lipgloss.Height(commandbarView)

	// Calculate content area height
	contentHeight := a.height - commandbarHeight

	var contentView string

	if a.compact {
		// Compact mode: header + page content
		headerView := a.header.View()
		headerHeight := lipgloss.Height(headerView)

		pageHeight := contentHeight - headerHeight
		a.activePage().SetSize(a.width, pageHeight)
		pageView := a.activePage().View()

		contentView = lipgloss.JoinVertical(
			lipgloss.Left,
			headerView,
			pageView,
		)
	} else {
		// Wide mode: page content + sidebar
		pageWidth := a.width - SidebarWidth
		a.activePage().SetSize(pageWidth, contentHeight)
		a.sidebar.SetSize(SidebarWidth, contentHeight)

		pageView := a.activePage().View()
		sidebarView := a.sidebar.View()

		contentView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			pageView,
			sidebarView,
		)
	}

	// Compose content + command bar
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		contentView,
		commandbarView,
	)

	// If we have popovers, layer them
	if len(a.popoverStack) > 0 {
		// TODO: Implement proper popover rendering
		// For now, just return the composed view
	}

	return lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Background(colors.Page.Bg).
		Render(view)
}

// SetSize sets dimensions and updates compact mode
func (a *App) SetSize(width, height int) {
	a.width = width
	a.height = height
	a.compact = width < CompactModeWidth

	// Update component sizes
	a.header.SetSize(width)
	a.commandbar.SetSize(width)

	// Page size is set in View() after calculating available space
}

// PushPopover adds a page to the popover stack
func (a *App) PushPopover(p page.Page) tea.Cmd {
	a.popoverStack = append(a.popoverStack, p)
	a.updateChrome()
	return p.Init()
}

// PopPopover removes the top popover
func (a *App) PopPopover() {
	if len(a.popoverStack) > 0 {
		a.popoverStack = a.popoverStack[:len(a.popoverStack)-1]
		a.updateChrome()
	}
}

// IsComplete returns false - app mode never completes
func (a *App) IsComplete() bool {
	return false
}

// IsBusy returns true if any layer is busy
func (a *App) IsBusy() bool {
	if a.chat.IsBusy() {
		return true
	}
	for _, p := range a.popoverStack {
		if p.IsBusy() {
			return true
		}
	}
	return false
}

// HasError returns true if the active layer has an error
func (a *App) HasError() bool {
	return a.activePage().HasError()
}

// Error returns the active layer's error
func (a *App) Error() error {
	return a.activePage().Error()
}
