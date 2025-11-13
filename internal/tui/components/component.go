package components

import tea "github.com/charmbracelet/bubbletea/v2"

// Component is the interface that all TUI components must implement.
//
// Components use pointer receivers for their methods. This follows standard Go
// practice for types that mutate state - methods that modify the receiver should
// use pointer receivers. This prevents the footgun of forgetting to capture a
// return value (which causes silent state loss with value receivers).
//
// The Update method does not return the component itself, only the command.
// This signature naturally encourages pointer receivers, since value receivers
// would mutate a copy and lose state.
type Component interface {
	// Init initializes the component and returns any initial command
	Init() tea.Cmd

	// Update handles a message and returns a command
	// The component is mutated in place via pointer receiver
	Update(tea.Msg) tea.Cmd

	// View renders the component to a string
	View() string

	// IsBusy returns true if the component is performing background work
	IsBusy() bool

	// HasError returns true if the component is in an error state
	HasError() bool

	// Error returns the current error, or nil if no error
	Error() error
}
