package layoutstest

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/tui/layouts"
)

// MockLayout implements layouts.Layout for testing
type MockLayout struct {
	UpdateFunc         func(msg tea.Msg) tea.Cmd
	SetSizeFunc        func(width, height int)
	SetKeyBindingsFunc func(bindings []key.Binding)
	SetErrorFunc       func(err error)
	ContentSizeFunc    func() (int, int)
	RenderFunc         func(content string) string

	// State for assertions
	LastError    error
	LastBindings []key.Binding
	LastWidth    int
	LastHeight   int
}

// Compile-time check that MockLayout implements layouts.Layout
var _ layouts.Layout = (*MockLayout)(nil)

// NewMockLayout creates a new mock layout with default implementations
func NewMockLayout() *MockLayout {
	return &MockLayout{
		ContentSizeFunc: func() (int, int) {
			return 80, 24
		},
		RenderFunc: func(content string) string {
			return content
		},
	}
}

func (m *MockLayout) Update(msg tea.Msg) tea.Cmd {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(msg)
	}
	return nil
}

func (m *MockLayout) SetSize(width, height int) {
	m.LastWidth = width
	m.LastHeight = height
	if m.SetSizeFunc != nil {
		m.SetSizeFunc(width, height)
	}
}

func (m *MockLayout) SetKeyBindings(bindings []key.Binding) {
	m.LastBindings = bindings
	if m.SetKeyBindingsFunc != nil {
		m.SetKeyBindingsFunc(bindings)
	}
}

func (m *MockLayout) SetError(err error) {
	m.LastError = err
	if m.SetErrorFunc != nil {
		m.SetErrorFunc(err)
	}
}

func (m *MockLayout) ContentSize() (int, int) {
	if m.ContentSizeFunc != nil {
		return m.ContentSizeFunc()
	}
	return 80, 24
}

func (m *MockLayout) Render(content string) string {
	if m.RenderFunc != nil {
		return m.RenderFunc(content)
	}
	return content
}
