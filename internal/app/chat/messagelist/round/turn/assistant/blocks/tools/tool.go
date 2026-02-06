package tools

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/thinking"
)

// State represents the current state of tool execution.
type State int

const (
	StateAccumulating State = iota
	StateExecuting
	StateComplete
)

// Status represents the outcome of a completed tool.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusError
)

// Icons for different statuses.
const (
	IconPending = "●"
	IconSuccess = "✓"
	IconError   = "×"
)

// Body padding: left=bodyPaddingLeft, right=bodyPaddingRight, top/bottom=1.
const (
	bodyPaddingLeft  = 2
	bodyPaddingRight = 1
	bodyPaddingH     = bodyPaddingLeft + bodyPaddingRight
)

// Child is the interface that specific tool models must implement.
type Child interface {
	Update(tea.Msg) tea.Cmd
	View() string
	SetWidth(int)
	Name() string
	Status() string // Message shown while executing (e.g., "Checking service status")
	Result() string // Message shown when complete (e.g., "Found 14 services")
	State() State
	ToolID() string
	Err() error
}

// Model is the chrome wrapper for tool blocks.
// It handles icon rendering, name display, animation, and content indentation.
// The actual tool logic lives in the embedded child.
// It is a fixed-height component.
type Model struct {
	theme  *styles.Theme
	index  int
	toolID string
	width  int
	status Status

	child    Child
	thinking *thinking.Model
}

// New creates a new tool model wrapping the given child.
func New(theme *styles.Theme, index int, toolID string, width int, child Child) *Model {
	// Child gets width minus body horizontal padding
	child.SetWidth(width - bodyPaddingH)
	return &Model{
		theme:    theme,
		index:    index,
		toolID:   toolID,
		width:    width,
		status:   StatusPending,
		child:    child,
		thinking: thinking.New(theme, thinking.Settings{Size: 10}),
	}
}

// Init starts the thinking animation.
func (m *Model) Init() tea.Cmd {
	return m.thinking.Init()
}

// Update handles messages - updates status and forwards to child.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update thinking animation
	cmds = append(cmds, m.thinking.Update(msg))

	// Update status based on child state changes
	m.updateStatus()

	// Listen for completion messages to update status
	if completed, ok := msg.(msgs.ToolCompleted); ok {
		if completed.GetToolUseID() == m.toolID {
			if completed.GetError() != nil {
				m.status = StatusError
			} else {
				m.status = StatusSuccess
			}
		}
	}

	// Forward to child
	cmds = append(cmds, m.child.Update(msg))

	return tea.Batch(cmds...)
}

// updateStatus syncs status with child state.
func (m *Model) updateStatus() {
	switch m.child.State() {
	case StateAccumulating:
		m.status = StatusPending
	case StateExecuting:
		m.status = StatusRunning
		// Update thinking label to show status message with reveal animation
		m.thinking.SetLabel(m.child.Status())
	case StateComplete:
		if m.child.Err() != nil {
			m.status = StatusError
		} else {
			m.status = StatusSuccess
		}
	}
}

// View renders the tool with chrome.
func (m *Model) View() string {
	colors := m.theme.Colors
	icon := m.renderIcon()
	nameStyle := lipgloss.NewStyle().Foreground(colors.Accent)
	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	switch m.status {
	case StatusPending:
		// ● Query ████████████
		return fmt.Sprintf("%s %s %s",
			icon,
			nameStyle.Render(m.child.Name()),
			m.thinking.View())

	case StatusRunning:
		// ● Query · Checking service status
		//
		//   ████████████
		header := fmt.Sprintf("%s %s", icon, nameStyle.Render(m.child.Name()))
		if status := m.child.Status(); status != "" {
			header = fmt.Sprintf("%s %s %s", icon, nameStyle.Render(m.child.Name()), mutedStyle.Render("· "+status))
		}
		body := lipgloss.NewStyle().
			PaddingLeft(bodyPaddingLeft).
			Render(m.thinking.View())
		return header + "\n\n" + body

	case StatusSuccess:
		// ✓ Query · Found 14 services
		//
		//   <child view if any>
		result := m.child.Result()
		header := fmt.Sprintf("%s %s", icon, nameStyle.Render(m.child.Name()))
		if result != "" {
			header = fmt.Sprintf("%s %s %s", icon, nameStyle.Render(m.child.Name()), mutedStyle.Render("· "+result))
		}
		childView := m.child.View()
		if childView == "" {
			return header
		}
		body := m.bodyStyle().Render(childView)
		return header + "\n\n" + body

	case StatusError:
		// × Query · Checking service status
		//
		//   ERROR  Connection timeout
		header := fmt.Sprintf("%s %s", icon, nameStyle.Render(m.child.Name()))
		if status := m.child.Status(); status != "" {
			header = fmt.Sprintf("%s %s %s", icon, nameStyle.Render(m.child.Name()), mutedStyle.Render("· "+status))
		}
		errTag := lipgloss.NewStyle().
			Background(colors.Error.Bg).
			Foreground(colors.Error.Fg).
			Padding(0, 1).
			Render("ERROR")
		errMsg := m.child.Err().Error()
		body := lipgloss.NewStyle().
			PaddingLeft(bodyPaddingLeft).
			Render(errTag + " " + mutedStyle.Render(errMsg))
		return header + "\n\n" + body
	}

	return ""
}

// Height returns the number of lines this block renders.
func (m *Model) Height() int {
	return lipgloss.Height(m.View())
}

// bodyStyle returns the style for the tool's content area.
func (m *Model) bodyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, bodyPaddingRight, 1, bodyPaddingLeft).
		Background(m.theme.Colors.Panel.Bg)
}

// renderIcon returns the colored status icon.
func (m *Model) renderIcon() string {
	colors := m.theme.Colors

	switch m.status {
	case StatusSuccess:
		return lipgloss.NewStyle().Foreground(colors.Success.Fg).Render(IconSuccess)
	case StatusError:
		return lipgloss.NewStyle().Foreground(colors.Error.Fg).Render(IconError)
	default:
		return lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render(IconPending)
	}
}

// ForceStatus sets the status directly (for testing).
func (m *Model) ForceStatus(status Status) {
	m.status = status
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.child.SetWidth(width - bodyPaddingH)
}

// Index returns the block index.
func (m *Model) Index() int {
	return m.index
}

// ToolID returns the tool ID.
func (m *Model) ToolID() string {
	return m.toolID
}

// Name returns the tool name.
func (m *Model) Name() string {
	return m.child.Name()
}

// State returns the child's state.
func (m *Model) State() State {
	return m.child.State()
}
