package page

import (
	"github.com/charmbracelet/bubbles/v2/help"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// Page represents a page within the app mode (chat, services, discovery, etc.)
// Pages are navigated via sidebar and can be switched at any time.
// Unlike onboarding steps which progress linearly, pages are random-access.
type Page interface {
	// Init initializes the page
	Init() tea.Cmd

	// Update handles messages and returns a command
	Update(tea.Msg) tea.Cmd

	// View returns the string representation of the page's UI
	View() string

	// SetSize sets the dimensions available for rendering
	SetSize(width, height int)

	// IsBusy returns true if the page is performing background work
	IsBusy() bool

	// HasError returns true if the page is in an error state
	HasError() bool

	// Error returns the current error, or nil if no error
	Error() error

	// Help returns the help key bindings for this page
	Help() help.KeyMap
}
