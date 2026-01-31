package app

import (
	"context"
	"sort"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	tuichat "github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/header"
	"github.com/usetero/cli/internal/tui/components/sidebar"
	"github.com/usetero/cli/internal/tui/layouts"
)

const (
	// Width threshold for switching between sidebar and header mode
	CompactModeWidth = 120

	// Sidebar width when in wide mode
	SidebarWidth = 32
)

// App is the main application orchestrator.
// It renders pages with appropriate chrome (sidebar or header) based on
// window size, uses Base layout for consistent padding and footer.
type App struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Layout - provides padding and footer
	layout *layouts.Base

	// Base layer - chat
	chat *tuichat.Chat

	// Popover stack - pages layered on top of chat
	popoverStack []page.Page

	// Chrome components (rendered inside layout)
	sidebar *sidebar.Sidebar
	header  *header.Header

	// Dependencies
	logger log.Logger
	db     sqlite.Database

	// Identity
	org     api.Organization
	account api.Account

	// Global key bindings (for footer display)
	globalBindings []key.Binding

	// Dimensions
	width  int
	height int

	// State
	compact bool // true when width < CompactModeWidth
}

// New creates a new app. Requires db for database access.
func New(ctx context.Context, theme *styles.Theme, db sqlite.Database, org api.Organization, account api.Account, logger log.Logger, globalBindings []key.Binding) *App {
	chatService := chat.NewService(db, logger)
	return &App{
		ctx:            ctx,
		theme:          theme,
		db:             db,
		layout:         layouts.NewBase(theme, logger),
		chat:           tuichat.New(theme, db, chatService, org.ID, account.ID, logger),
		sidebar:        sidebar.New(theme, logger),
		header:         header.New(theme, logger),
		logger:         logger,
		org:            org,
		account:        account,
		globalBindings: globalBindings,
	}
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	return a.chat.Init()
}

// activePage returns the topmost active page (popover or chat).
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
	case tea.KeyPressMsg:
		// Check for popover dismiss (Esc)
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) && len(a.popoverStack) > 0 {
			a.popoverStack = a.popoverStack[:len(a.popoverStack)-1]
			a.updateChrome()
			return nil
		}
	}

	// Update layout (handles footer)
	cmd := a.layout.Update(msg)
	cmds = append(cmds, cmd)

	// Route to active page
	cmd = a.activePage().Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// updateChrome updates sidebar/header/footer with current page's metadata
func (a *App) updateChrome() {
	p := a.activePage()
	if p == nil {
		// No page yet, set defaults
		a.sidebar.SetTitle("Tero")
		a.sidebar.SetOrgName(a.org.Name)
		a.header.SetTitle("Tero")
		a.header.SetOrgName(a.org.Name)
		a.layout.SetKeyBindings(a.globalBindings)
		return
	}

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

	// Update footer with keybindings
	allBindings := append(p.KeyBindings(), a.globalBindings...)
	a.layout.SetKeyBindings(allBindings)
}

// View renders the app
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return ""
	}

	a.updateChrome()

	// Get content dimensions from layout
	contentWidth, contentHeight := a.layout.ContentSize()

	var contentView string

	if a.compact {
		// Compact mode: header + page content
		headerView := a.header.View()
		headerHeight := lipgloss.Height(headerView)

		pageHeight := contentHeight - headerHeight
		a.activePage().SetSize(contentWidth, pageHeight)
		pageView := a.activePage().View()

		contentView = lipgloss.JoinVertical(
			lipgloss.Left,
			headerView,
			pageView,
		)
	} else {
		// Wide mode: page content + sidebar
		pageWidth := contentWidth - SidebarWidth
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

	// Wrap in layout (adds padding + footer)
	return a.layout.Render(contentView)
}

// SetSize sets dimensions and updates compact mode
func (a *App) SetSize(width, height int) {
	a.width = width
	a.height = height
	a.compact = width < CompactModeWidth

	// Update layout size
	a.layout.SetSize(width, height)

	// Update header width (after layout padding)
	contentWidth, _ := a.layout.ContentSize()
	a.header.SetSize(contentWidth)
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

// HasError returns true if active layer has an error
func (a *App) HasError() bool {
	return a.activePage().HasError()
}

// Error returns the current error
func (a *App) Error() error {
	return a.activePage().Error()
}
