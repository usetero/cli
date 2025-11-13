package app

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/page"
)

// App represents the app mode - the main application with sidebar navigation.
// It manages pages (chat, services, discovery, settings) and handles sidebar routing.
type App struct {
	currentPage    page.Page
	logger         log.Logger
	orgID          string
	accountID      string
	globalBindings []key.Binding
}

// New creates a new app mode starting with the chat page
func New(orgID string, accountID string, logger log.Logger, globalBindings []key.Binding) *App {
	// Create chat page as the initial page
	chatPage := chat.New(orgID, accountID, logger, globalBindings)

	return &App{
		currentPage:    chatPage,
		logger:         logger,
		orgID:          orgID,
		accountID:      accountID,
		globalBindings: globalBindings,
	}
}

// Init initializes the app mode
func (m *App) Init() tea.Cmd {
	return m.currentPage.Init()
}

// Update handles messages and delegates to the current page
func (m *App) Update(msg tea.Msg) tea.Cmd {
	// Future: Handle sidebar navigation here
	// For now, just delegate to current page
	return m.currentPage.Update(msg)
}

// View renders the current page
func (m *App) View() string {
	return m.currentPage.View()
}

// SetSize sets dimensions and propagates to current page
func (m *App) SetSize(width, height int) {
	m.currentPage.SetSize(width, height)
}

// IsComplete returns false - app mode never completes
func (m *App) IsComplete() bool {
	return false
}

// IsBusy delegates to the current page
func (m *App) IsBusy() bool {
	return m.currentPage.IsBusy()
}

// HasError returns true if the current page has an error
func (m *App) HasError() bool {
	if m.currentPage == nil {
		return false
	}
	return m.currentPage.HasError()
}

// Error returns the current page's error, or nil if no error
func (m *App) Error() error {
	if m.currentPage == nil {
		return nil
	}
	return m.currentPage.Error()
}
