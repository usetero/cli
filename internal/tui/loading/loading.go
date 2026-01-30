package loading

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/sync"
)

// Loading is the mode shown while waiting for initial sync to complete.
// It displays a loading indicator and transitions to App once sync is done.
type Loading struct {
	theme *styles.Theme

	// Identity (passed through to App)
	org     api.Organization
	account api.Account

	// State
	db     sqlite.Database // set when sync completes
	err    error
	width  int
	height int
}

// New creates a new loading mode.
func New(theme *styles.Theme, org api.Organization, account api.Account) *Loading {
	return &Loading{
		theme:   theme,
		org:     org,
		account: account,
	}
}

// Init initializes the loading mode.
func (l *Loading) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (l *Loading) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case sync.CompletedMsg:
		l.db = msg.DB
	}
	return nil
}

// View renders the loading screen.
func (l *Loading) View() string {
	if l.width == 0 || l.height == 0 {
		return ""
	}

	colors := l.theme.Colors

	var content string
	if l.err != nil {
		content = lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Render("Sync error: " + l.err.Error())
	} else {
		content = lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render("Connecting...")
	}

	return lipgloss.NewStyle().
		Width(l.width).
		Height(l.height).
		Background(colors.Page.Bg).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// SetSize sets the dimensions.
func (l *Loading) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// IsComplete returns true when sync has completed successfully.
func (l *Loading) IsComplete() bool {
	return l.db != nil
}

// IsBusy returns true while waiting for sync.
func (l *Loading) IsBusy() bool {
	return l.db == nil && l.err == nil
}

// HasError returns true if sync failed.
func (l *Loading) HasError() bool {
	return l.err != nil
}

// Error returns the sync error.
func (l *Loading) Error() error {
	return l.err
}

// DB returns the database (only valid after IsComplete returns true).
func (l *Loading) DB() sqlite.Database {
	return l.db
}

// Organization returns the organization.
func (l *Loading) Organization() api.Organization {
	return l.org
}

// Account returns the account.
func (l *Loading) Account() api.Account {
	return l.account
}
