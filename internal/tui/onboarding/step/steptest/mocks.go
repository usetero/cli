package steptest

import (
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// MockStep implements step.Step for testing
type MockStep struct {
	InitFunc       func() tea.Cmd
	UpdateFunc     func(tea.Msg) (step.Step, tea.Cmd)
	ViewFunc       func() string
	SetSizeFunc    func(width, height int)
	IsCompleteFunc func() bool
	IsBusyFunc     func() bool
	HasErrorFunc   func() bool
	ErrorFunc      func() error
	HelpFunc       func() help.KeyMap
	NextFunc       func() step.Step

	// State for testing
	Err error
}

// Compile-time check that MockStep implements step.Step
var _ step.Step = (*MockStep)(nil)

// NewMockStep creates a new mock step with default implementations
func NewMockStep() *MockStep {
	return &MockStep{}
}

func (m *MockStep) Init() tea.Cmd {
	if m.InitFunc != nil {
		return m.InitFunc()
	}
	return nil
}

func (m *MockStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(msg)
	}
	return m, nil
}

func (m *MockStep) View() string {
	if m.ViewFunc != nil {
		return m.ViewFunc()
	}
	return "mock step"
}

func (m *MockStep) SetSize(width, height int) {
	if m.SetSizeFunc != nil {
		m.SetSizeFunc(width, height)
	}
}

func (m *MockStep) IsComplete() bool {
	if m.IsCompleteFunc != nil {
		return m.IsCompleteFunc()
	}
	return false
}

func (m *MockStep) IsBusy() bool {
	if m.IsBusyFunc != nil {
		return m.IsBusyFunc()
	}
	return false
}

func (m *MockStep) HasError() bool {
	if m.HasErrorFunc != nil {
		return m.HasErrorFunc()
	}
	return m.Err != nil
}

func (m *MockStep) Error() error {
	if m.ErrorFunc != nil {
		return m.ErrorFunc()
	}
	return m.Err
}

func (m *MockStep) Help() help.KeyMap {
	if m.HelpFunc != nil {
		return m.HelpFunc()
	}
	return keymap.Simple{Keys: []key.Binding{}}
}

func (m *MockStep) Next() step.Step {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	return nil
}
