package tools

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
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
	StatusCancelled
)

// Icons for different statuses.
const (
	IconPending   = "●"
	IconSuccess   = "✓"
	IconError     = "×"
	IconCancelled = "○"
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
// It is a fixed-height component. Implements block.Block.
type Model struct {
	theme  styles.Theme
	index  int
	toolID string
	width  int
	status Status

	child    Child
	thinking *thinking.Model
	focused  bool
	expanded bool // whether the body content is visible (only applies to completed tools)
}

// New creates a new tool model wrapping the given child.
func New(theme styles.Theme, index int, toolID string, width int, child Child) *Model {
	// Child gets width minus outer padding and body padding
	child.SetWidth(width - block.PaddingX*2 - bodyPaddingH)
	return &Model{
		theme:    theme,
		index:    index,
		toolID:   toolID,
		width:    width,
		status:   StatusPending,
		child:    child,
		thinking: thinking.New(theme, thinking.Settings{Size: 10}),
		expanded: false,
	}
}

// Init starts the thinking animation.
func (m *Model) Init() tea.Cmd {
	return m.thinking.Init()
}

// Cancel stops the tool's thinking animation and marks it cancelled.
func (m *Model) Cancel() {
	m.status = StatusCancelled
}

// Update handles messages - updates status and forwards to child.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Only tick the thinking animation while still pending/running
	if m.status == StatusPending || m.status == StatusRunning {
		cmds = append(cmds, m.thinking.Update(msg))
	}

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
	colors := m.theme
	icon := m.renderIcon()
	nameStyle := lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg)
	mutedStyle := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sp := mutedStyle.Render(" ")

	var content string

	switch m.status {
	case StatusPending:
		// ● Query ████████████
		content = icon + sp + nameStyle.Render(m.child.Name()) + sp + m.thinking.View()

	case StatusRunning:
		// ● Query · Checking service status
		//
		//   ████████████
		header := icon + sp + nameStyle.Render(m.child.Name())
		if status := m.child.Status(); status != "" {
			header = icon + sp + nameStyle.Render(m.child.Name()) + sp + mutedStyle.Render("· "+status)
		}
		body := lipgloss.NewStyle().
			PaddingLeft(bodyPaddingLeft).
			Render(m.thinking.View())
		content = header + "\n\n" + body

	case StatusSuccess:
		// ✓ ▶ Query · Found 14 services  (collapsed)
		// ✓ ▼ Query · Found 14 services  (expanded)
		//
		//   <child view if any>
		result := m.child.Result()
		chevron := mutedStyle.Render("▶")
		if m.expanded {
			chevron = mutedStyle.Render("▼")
		}
		header := icon + sp + chevron + sp + nameStyle.Render(m.child.Name())
		if result != "" {
			header = icon + sp + chevron + sp + nameStyle.Render(m.child.Name()) + sp + mutedStyle.Render("· "+result)
		}
		if !m.expanded || m.child.View() == "" {
			content = header
		} else {
			body := m.bodyStyle().Render(m.child.View())
			content = header + "\n\n" + body
		}

	case StatusError:
		// × Query · Checking service status
		//
		//   ERROR  Connection timeout
		header := icon + sp + nameStyle.Render(m.child.Name())
		if status := m.child.Status(); status != "" {
			header = icon + sp + nameStyle.Render(m.child.Name()) + sp + mutedStyle.Render("· "+status)
		}
		errTag := lipgloss.NewStyle().
			Background(colors.Error).
			Foreground(colors.OnError).
			Padding(0, 1).
			Render("ERROR")
		errMsg := m.child.Err().Error()
		body := lipgloss.NewStyle().
			PaddingLeft(bodyPaddingLeft).
			Render(errTag + sp + mutedStyle.Render(errMsg))
		content = header + "\n\n" + body

	case StatusCancelled:
		// ○ Query
		content = icon + sp + nameStyle.Render(m.child.Name())
	}

	return lipgloss.NewStyle().
		Background(colors.Bg).
		Padding(block.PaddingY, block.PaddingX).
		Width(m.width).
		Render(content)
}

// Height returns the number of lines this block renders.
func (m *Model) Height() int {
	return lipgloss.Height(m.View())
}

// bodyStyle returns the style for the tool's content area.
func (m *Model) bodyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, bodyPaddingRight, 1, bodyPaddingLeft).
		Background(m.theme.Bg)
}

// renderIcon returns the colored status icon.
func (m *Model) renderIcon() string {
	colors := m.theme
	bg := colors.Bg

	switch m.status {
	case StatusSuccess:
		return lipgloss.NewStyle().Foreground(colors.Success).Background(bg).Render(IconSuccess)
	case StatusError:
		return lipgloss.NewStyle().Foreground(colors.Error).Background(bg).Render(IconError)
	case StatusCancelled:
		return lipgloss.NewStyle().Foreground(colors.TextMuted).Background(bg).Render(IconCancelled)
	default:
		return lipgloss.NewStyle().Foreground(colors.TextMuted).Background(bg).Render(IconPending)
	}
}

// ForceStatus sets the status directly (for testing).
func (m *Model) ForceStatus(status Status) {
	m.status = status
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.child.SetWidth(width - block.PaddingX*2 - bodyPaddingH)
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

// Kind implements block.Block.
func (m *Model) Kind() block.Kind {
	return block.KindTool
}

// SetFocused implements block.Block.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// Focused implements block.Block.
func (m *Model) Focused() bool {
	return m.focused
}

// Toggle implements block.Toggleable — expands/collapses the tool body.
func (m *Model) Toggle() {
	m.expanded = !m.expanded
}
