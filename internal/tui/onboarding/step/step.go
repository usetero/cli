package step

import (
	"errors"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

// ErrNotReady is returned by Next() when the step is not yet complete.
var ErrNotReady = errors.New("step not ready")

// Step represents a single step in the onboarding flow.
// Implementations must use value receivers and return new values.
type Step interface {
	// Init initializes the step and returns any initial commands
	Init() tea.Cmd

	// Update handles messages and returns the updated step
	Update(tea.Msg) (Step, tea.Cmd)

	// View renders the step's UI
	View() string

	// SetSize returns a new step with the given dimensions
	SetSize(width, height int) Step

	// IsBusy returns true if performing a background operation
	IsBusy() bool

	// HasError returns true if in an error state
	HasError() bool

	// Error returns the current error, or nil
	Error() error

	// Help returns the key bindings for this step
	Help() help.KeyMap

	// Next returns the next step in the flow.
	// Returns:
	//   - (nextStep, nil) - transition to nextStep
	//   - (nil, nil) - flow complete
	//   - (nil, ErrNotReady) - not complete yet
	//   - (nil, err) - step failed
	Next() (Step, error)

	// Close releases any resources
	Close() error
}
