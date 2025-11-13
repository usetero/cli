package step

import (
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/tui/keymap"
)

// Flow orchestrates a chain of steps.
// Steps transition automatically when complete by calling Next() to get the next step.
// Uses pointer receiver pattern for efficiency.
type Flow struct {
	current Step
	width   int
	height  int
}

// NewFlow creates a new flow starting with the given step
func NewFlow(startStep Step) *Flow {
	return &Flow{
		current: startStep,
	}
}

// Init initializes the current step
func (f *Flow) Init() tea.Cmd {
	if f.current == nil {
		return nil
	}
	return f.current.Init()
}

// Update handles messages and auto-transitions when steps complete
func (f *Flow) Update(msg tea.Msg) tea.Cmd {
	if f.current == nil {
		return nil
	}

	// Update current step
	var cmd tea.Cmd
	f.current, cmd = f.current.Update(msg)

	// Auto-transition when current step completes
	if f.current.IsComplete() {
		// Transition to next step, skipping any that are already complete
		// (e.g., from saved preferences). This loop handles chains of
		// pre-satisfied steps gracefully.
		for f.current.IsComplete() {
			nextStep := f.current.Next()
			if nextStep == nil {
				// No more steps - flow complete
				f.current = nil
				return cmd
			}

			f.current = nextStep
			f.current.SetSize(f.width, f.height)
			initCmd := f.current.Init()
			cmd = tea.Batch(cmd, initCmd)
		}
	}

	return cmd
}

// View renders the current step
func (f *Flow) View() string {
	if f.current == nil {
		return ""
	}
	return f.current.View()
}

// SetSize sets the size for the current step
func (f *Flow) SetSize(width, height int) {
	f.width = width
	f.height = height
	if f.current != nil {
		f.current.SetSize(width, height)
	}
}

// IsComplete returns true if flow has no more steps (current is nil)
func (f *Flow) IsComplete() bool {
	return f.current == nil
}

// IsBusy returns true if the current step is busy
func (f *Flow) IsBusy() bool {
	if f.current == nil {
		return false
	}
	return f.current.IsBusy()
}

// HasError returns true if the current step has an error
func (f *Flow) HasError() bool {
	if f.current == nil {
		return false
	}
	return f.current.HasError()
}

// Error returns the current step's error, or nil if no error
func (f *Flow) Error() error {
	if f.current == nil {
		return nil
	}
	return f.current.Error()
}

// Current returns the current step (for debugging/inspection)
func (f *Flow) Current() Step {
	return f.current
}

// Help delegates to the current step's help
func (f *Flow) Help() help.KeyMap {
	if f.current == nil {
		return keymap.Simple{Keys: []key.Binding{}}
	}
	return f.current.Help()
}
