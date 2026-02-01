package step

import (
	"errors"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

// ErrNotReady is returned by Next() when the step is not yet complete.
var ErrNotReady = errors.New("step not ready")

// Step represents a single step in the onboarding flow.
// Steps are self-contained components that collect user input and pass data forward.
//
// Each step manages its own layout and state.
// The onboarding flow orchestrates progression through steps.
type Step interface {
	// Init initializes the step and returns any initial commands
	Init() tea.Cmd

	// Update handles messages and updates step state
	// Returns the updated step and any commands to execute
	Update(tea.Msg) (Step, tea.Cmd)

	// View renders the step's UI as a string
	View() string

	// SetSize sets the width and height available for rendering
	SetSize(width, height int)

	// IsBusy returns true if this step is performing a background operation
	// (e.g., network request, validation, etc.)
	IsBusy() bool

	// HasError returns true if this step is in an error state
	HasError() bool

	// Error returns the current error, or nil if no error
	Error() error

	// Help returns the key bindings for this step
	Help() help.KeyMap

	// Next returns the next step in the flow.
	// Returns:
	//   - (nextStep, nil) - transition to nextStep
	//   - (nil, nil) - flow complete, no more steps
	//   - (nil, ErrNotReady) - step not complete yet, stay on current step
	//   - (nil, err) - step failed with error
	Next() (Step, error)

	// Close releases any resources held by the step.
	Close() error
}
