package loading

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// SyncState provides read-only access to sync manager state.
type SyncState interface {
	DB() sqlite.Database
}

// Loading is the mode shown while waiting for initial sync to complete.
// It displays a loading indicator and transitions to App once sync is done.
type Loading struct {
	theme *styles.Theme

	// Identity (passed through to App)
	org     api.Organization
	account api.Account

	// Sync state - checked directly rather than relying on messages
	syncState SyncState

	// State
	width  int
	height int
}

// New creates a new loading mode.
func New(theme *styles.Theme, org api.Organization, account api.Account, syncState SyncState) *Loading {
	return &Loading{
		theme:     theme,
		org:       org,
		account:   account,
		syncState: syncState,
	}
}

// Init initializes the loading mode.
func (l *Loading) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (l *Loading) Update(msg tea.Msg) tea.Cmd {
	// No message handling needed - we check syncState directly
	_ = msg
	return nil
}

// View renders the loading screen.
func (l *Loading) View() string {
	if l.width == 0 || l.height == 0 {
		return ""
	}

	colors := l.theme.Colors

	content := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Render("Connecting...")

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
	return l.syncState.DB() != nil
}

// IsBusy returns true while waiting for sync.
func (l *Loading) IsBusy() bool {
	return l.syncState.DB() == nil
}

// HasError returns true if sync failed.
func (l *Loading) HasError() bool {
	return false // TODO: expose error from sync manager if needed
}

// Error returns the sync error.
func (l *Loading) Error() error {
	return nil // TODO: expose error from sync manager if needed
}

// DB returns the database (only valid after IsComplete returns true).
func (l *Loading) DB() sqlite.Database {
	return l.syncState.DB()
}

// Organization returns the organization.
func (l *Loading) Organization() api.Organization {
	return l.org
}

// Account returns the account.
func (l *Loading) Account() api.Account {
	return l.account
}
