// Package toast provides a toast notification component for the app.
package toast

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/styles"
)

// DefaultTTL is how long a toast shows before auto-dismissing.
const DefaultTTL = 5 * time.Second

// toastType represents the kind of toast being displayed.
type toastType int

const (
	toastError toastType = iota
	toastWarning
	toastSuccess
	toastInfo
)

// clearMsg dismisses the current toast.
type clearMsg struct{}

// toast holds the current toast state.
type toast struct {
	typ     toastType
	content string
	sticky  bool
}

// Model holds the toast state and renders it.
// It is a fixed-height component - height is 0 when no toast, 1 when showing.
type Model struct {
	theme   styles.Theme
	current *toast
	width   int
}

// New creates a new toast model.
func New(theme styles.Theme) *Model {
	return &Model{
		theme: theme,
	}
}

// SetWidth sets the available width for rendering.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// Height returns the number of lines this component renders.
// Always returns 1 to reserve space and prevent layout shifts.
func (m *Model) Height() int {
	return 1
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case msgs.Error:
		return m.show(toast{typ: toastError, content: msg.Message, sticky: msg.Sticky})
	case msgs.Warning:
		return m.show(toast{typ: toastWarning, content: msg.Message, sticky: msg.Sticky})
	case msgs.Success:
		return m.show(toast{typ: toastSuccess, content: msg.Message})
	case msgs.Info:
		return m.show(toast{typ: toastInfo, content: msg.Message})
	case clearMsg:
		m.current = nil
		return nil
	}
	return nil
}

// show displays a toast. Returns a command to auto-clear if not sticky.
func (m *Model) show(t toast) tea.Cmd {
	m.current = &t
	if t.sticky {
		return nil
	}
	return clearAfter(DefaultTTL)
}

// View renders the toast. Returns an empty line if no toast to reserve space.
func (m *Model) View() string {
	if m.current == nil {
		return " "
	}

	t := m.theme
	var labelStyle, msgStyle lipgloss.Style
	var label string

	switch m.current.typ {
	case toastError:
		label = "ERROR"
		labelStyle = lipgloss.NewStyle().
			Background(t.Error).
			Foreground(t.OnError).
			Padding(0, 1).
			Bold(true)
		msgStyle = lipgloss.NewStyle().
			Background(t.Error).
			Foreground(t.OnError).
			Padding(0, 1)
	case toastWarning:
		label = "WARNING"
		labelStyle = lipgloss.NewStyle().
			Background(t.Warning).
			Foreground(t.OnWarning).
			Padding(0, 1).
			Bold(true)
		msgStyle = lipgloss.NewStyle().
			Background(t.Warning).
			Foreground(t.OnWarning).
			Padding(0, 1)
	case toastSuccess:
		label = "OK"
		labelStyle = lipgloss.NewStyle().
			Background(t.Success).
			Foreground(t.OnSuccess).
			Padding(0, 1).
			Bold(true)
		msgStyle = lipgloss.NewStyle().
			Background(t.Success).
			Foreground(t.OnSuccess).
			Padding(0, 1)
	case toastInfo:
		label = "INFO"
		labelStyle = lipgloss.NewStyle().
			Background(t.Info).
			Foreground(t.OnInfo).
			Padding(0, 1).
			Bold(true)
		msgStyle = lipgloss.NewStyle().
			Background(t.Info).
			Foreground(t.OnInfo).
			Padding(0, 1)
	}

	labelText := labelStyle.Render(label)
	labelWidth := lipgloss.Width(labelText)

	msgWidth := m.width - labelWidth
	if msgWidth < 0 {
		msgWidth = 0
	}

	content := ansi.Truncate(m.current.content, msgWidth-2, "…") // -2 for padding
	msgText := msgStyle.Width(msgWidth).Render(content)

	return labelText + msgText
}

// clearAfter returns a command that sends clearMsg after the duration.
func clearAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMsg{}
	})
}
