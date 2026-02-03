package step

import (
	"errors"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/keymap"
)

// Flow orchestrates a chain of steps.
// Uses value receivers - all methods return new Flow values.
type Flow struct {
	current  Step
	lastStep Step
	logger   log.Logger
	width    int
	height   int
}

// NewFlow creates a new flow starting with the given step.
func NewFlow(startStep Step, logger log.Logger) Flow {
	return Flow{
		current: startStep,
		logger:  logger,
	}
}

// Init initializes the current step.
func (f Flow) Init() tea.Cmd {
	if f.current == nil {
		return nil
	}
	return f.current.Init()
}

// Update handles messages and auto-transitions when steps complete.
func (f Flow) Update(msg tea.Msg) (Flow, tea.Cmd) {
	if f.current == nil {
		return f, nil
	}

	// Update current step
	var cmd tea.Cmd
	f.current, cmd = f.current.Update(msg)

	// Try to transition to next step
	for {
		nextStep, err := f.current.Next()
		if errors.Is(err, ErrNotReady) {
			break
		}
		if err != nil {
			break
		}
		if nextStep == nil {
			// Flow complete
			f.lastStep = f.current
			if err := f.current.Close(); err != nil {
				f.logger.Error("failed to close step", "error", err)
			}
			f.current = nil
			return f, cmd
		}

		// Transition to next step
		if err := f.current.Close(); err != nil {
			f.logger.Error("failed to close step", "error", err)
		}
		f.current = nextStep.SetSize(f.width, f.height)
		initCmd := f.current.Init()
		cmd = tea.Batch(cmd, initCmd)
	}

	return f, cmd
}

// View renders the current step.
func (f Flow) View() string {
	if f.current == nil {
		return ""
	}
	return f.current.View()
}

// SetSize returns a new Flow with the given dimensions.
func (f Flow) SetSize(width, height int) Flow {
	f.width = width
	f.height = height
	if f.current != nil {
		f.current = f.current.SetSize(width, height)
	}
	return f
}

// IsComplete returns true if flow has no more steps.
func (f Flow) IsComplete() bool {
	return f.current == nil
}

// IsBusy returns true if the current step is busy.
func (f Flow) IsBusy() bool {
	if f.current == nil {
		return false
	}
	return f.current.IsBusy()
}

// HasError returns true if the current step has an error.
func (f Flow) HasError() bool {
	if f.current == nil {
		return false
	}
	return f.current.HasError()
}

// Error returns the current step's error.
func (f Flow) Error() error {
	if f.current == nil {
		return nil
	}
	return f.current.Error()
}

// Current returns the current step.
func (f Flow) Current() Step {
	return f.current
}

// LastStep returns the final step before flow completed.
func (f Flow) LastStep() Step {
	return f.lastStep
}

// Help returns the current step's help.
func (f Flow) Help() help.KeyMap {
	if f.current == nil {
		return keymap.Simple{Keys: []key.Binding{}}
	}
	return f.current.Help()
}

// Close releases any resources held by the current step.
func (f Flow) Close() error {
	if f.current != nil {
		return f.current.Close()
	}
	return nil
}
