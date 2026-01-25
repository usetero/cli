package page

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Page represents a page within the app (chat, services, policies, etc.)
// Pages expose their content and capabilities. App handles all chrome
// (sidebar, header, command bar) based on these capabilities.
type Page interface {
	// Init initializes the page
	Init() tea.Cmd

	// Update handles messages and returns a command
	Update(tea.Msg) tea.Cmd

	// View returns the page content (no chrome - app adds that)
	View() string

	// SetSize sets the dimensions available for content rendering
	SetSize(width, height int)

	// Identity

	// Title returns the page title for the header (e.g., "Services", "Policies")
	Title() string

	// Metadata returns information to display in sidebar/header
	// Items are sorted by Priority (lower = shown first)
	Metadata() []Metadata

	// Capabilities

	// AcceptsNaturalLanguage returns true if the page accepts free-form text
	// input that goes to the AI (includes @references, questions, etc.)
	AcceptsNaturalLanguage() bool

	// Commands returns the slash commands this page supports
	Commands() []Command

	// KeyBindings returns keyboard shortcuts for this page (shown in footer)
	KeyBindings() []key.Binding

	// State

	// IsBusy returns true if the page is performing background work
	IsBusy() bool

	// HasError returns true if the page is in an error state
	HasError() bool

	// Error returns the current error, or nil if no error
	Error() error
}
