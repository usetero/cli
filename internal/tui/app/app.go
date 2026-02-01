package app

import (
	"context"
	"sort"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	tuichat "github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/commandbar"
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
//
// App owns the command bar and conversation state. You are always in a
// conversation (though it may not exist in SQLite until the first message).
type App struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Layout - provides padding and footer
	layout *layouts.Base

	// Command bar - always visible at bottom, owned by App
	commandBar *commandbar.CommandBar

	// Current conversation ID (empty until first message sent)
	conversationID string

	// Chat page - renders messages for current conversation
	chat *tuichat.Chat

	// Focus stack - pages layered on top of chat (detours)
	focusStack []page.Page

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
	return &App{
		ctx:            ctx,
		theme:          theme,
		db:             db,
		layout:         layouts.NewBase(theme, logger),
		commandBar:     commandbar.New(theme, logger),
		chat:           tuichat.New(theme, db, logger),
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
	return tea.Batch(
		a.commandBar.Init(),
		a.chat.Init(),
	)
}

// focusedPage returns the topmost focused page (detour or chat).
func (a *App) focusedPage() page.Page {
	if len(a.focusStack) > 0 {
		return a.focusStack[len(a.focusStack)-1]
	}
	return a.chat
}

// messageSentMsg is sent when a message has been written to SQLite.
type messageSentMsg struct {
	conversationID string
	err            error
}

// Update handles messages and routes to the appropriate layer
func (a *App) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Check for focus stack dismiss (Esc)
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) && len(a.focusStack) > 0 {
			a.focusStack = a.focusStack[:len(a.focusStack)-1]
			a.updateChrome()
			return nil
		}

	case commandbar.SubmitMsg:
		// User submitted input - send message
		return a.sendMessage(msg.Text)

	case messageSentMsg:
		if msg.err != nil {
			a.logger.Error("failed to send message", "error", msg.err)
			return nil
		}
		// Update conversation ID if this was the first message
		if a.conversationID == "" && msg.conversationID != "" {
			a.conversationID = msg.conversationID
			return a.chat.SetConversation(msg.conversationID)
		}
		return nil
	}

	// Update layout (handles footer)
	cmd := a.layout.Update(msg)
	cmds = append(cmds, cmd)

	// Update command bar (always visible)
	cmd = a.commandBar.Update(msg)
	cmds = append(cmds, cmd)

	// Route to focused page
	cmd = a.focusedPage().Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// sendMessage writes a user message directly to SQLite.
// The upload loop will sync it to the backend.
func (a *App) sendMessage(text string) tea.Cmd {
	return func() tea.Msg {
		now := time.Now().UTC().Format(time.RFC3339)
		convID := a.conversationID

		// Create conversation if needed
		if convID == "" {
			convID = uuid.New().String()
			accountID := a.account.ID
			err := a.db.Queries().InsertConversation(a.ctx, sqlite.InsertConversationParams{
				ID:        &convID,
				AccountID: &accountID,
				CreatedAt: &now,
				UpdatedAt: &now,
			})
			if err != nil {
				return messageSentMsg{err: err}
			}
		}

		// Insert the message
		msgID := uuid.New().String()
		accountID := a.account.ID
		role := "user"
		err := a.db.Queries().InsertMessage(a.ctx, sqlite.InsertMessageParams{
			ID:             &msgID,
			AccountID:      &accountID,
			ConversationID: &convID,
			Content:        &text,
			CreatedAt:      &now,
			Role:           &role,
		})
		if err != nil {
			return messageSentMsg{err: err}
		}

		return messageSentMsg{conversationID: convID, err: nil}
	}
}

// updateChrome updates sidebar/header/footer with current page's metadata
func (a *App) updateChrome() {
	p := a.focusedPage()
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

	// Command bar is always visible at the bottom
	commandBarView := a.commandBar.View()
	commandBarHeight := a.commandBar.Height()

	// Remaining height for page content
	pageAreaHeight := contentHeight - commandBarHeight

	var pageAreaView string

	if a.compact {
		// Compact mode: header + page content
		headerView := a.header.View()
		headerHeight := lipgloss.Height(headerView)

		pageHeight := pageAreaHeight - headerHeight
		a.focusedPage().SetSize(contentWidth, pageHeight)
		pageView := a.focusedPage().View()

		pageAreaView = lipgloss.JoinVertical(
			lipgloss.Left,
			headerView,
			pageView,
		)
	} else {
		// Wide mode: page content + sidebar
		pageWidth := contentWidth - SidebarWidth
		a.focusedPage().SetSize(pageWidth, pageAreaHeight)
		a.sidebar.SetSize(SidebarWidth, pageAreaHeight)

		pageView := a.focusedPage().View()
		sidebarView := a.sidebar.View()

		pageAreaView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			pageView,
			sidebarView,
		)
	}

	// Compose: page area + command bar
	contentView := lipgloss.JoinVertical(
		lipgloss.Left,
		pageAreaView,
		commandBarView,
	)

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

	// Update component widths (after layout padding)
	contentWidth, _ := a.layout.ContentSize()
	a.header.SetSize(contentWidth)
	a.commandBar.SetSize(contentWidth)
}

// PushFocus adds a page to the focus stack (a detour).
func (a *App) PushFocus(p page.Page) tea.Cmd {
	a.focusStack = append(a.focusStack, p)
	a.updateChrome()
	return p.Init()
}

// PopFocus removes the top page from the focus stack.
func (a *App) PopFocus() {
	if len(a.focusStack) > 0 {
		a.focusStack = a.focusStack[:len(a.focusStack)-1]
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
	for _, p := range a.focusStack {
		if p.IsBusy() {
			return true
		}
	}
	return false
}

// HasError returns true if focused page has an error
func (a *App) HasError() bool {
	return a.focusedPage().HasError()
}

// Error returns the current error
func (a *App) Error() error {
	return a.focusedPage().Error()
}
