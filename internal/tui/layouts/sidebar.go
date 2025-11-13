package layouts

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/sidebar"
)

const (
	SidebarWidth = 35 // Width of sidebar on the left
)

// Sidebar is a layout with a sidebar on the left and content on the right.
// Wraps everything in Base layout for footer and padding.
type Sidebar struct {
	// Nested layouts
	base *Base

	// Child components
	sidebar sidebar.Component

	// State
	width  int
	height int
}

// NewSidebar creates a new sidebar layout
func NewSidebar(logger log.Logger) *Sidebar {
	return &Sidebar{
		base:    NewBase(logger),
		sidebar: sidebar.New(logger),
	}
}

// Update handles messages for the sidebar layout
func (s *Sidebar) Update(msg tea.Msg) tea.Cmd {
	// Cascade to base
	return s.base.Update(msg)
}

// SetSize sets the dimensions for the layout
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height

	// Base gets full dimensions
	s.base.SetSize(width, height)

	// Sidebar gets fixed width and height after accounting for base padding
	_, baseContentHeight := s.base.ContentSize()
	s.sidebar.SetSize(SidebarWidth, baseContentHeight)
}

// SetKeyBindings updates the key bindings shown in the footer
func (s *Sidebar) SetKeyBindings(bindings []key.Binding) {
	s.base.SetKeyBindings(bindings)
}

// SetError sets the error to display in the footer
func (s *Sidebar) SetError(err error) {
	s.base.SetError(err)
}

// ContentSize returns the available space for content (width x height after sidebar and footer)
func (s *Sidebar) ContentSize() (int, int) {
	if s.width == 0 || s.height == 0 {
		return 0, 0
	}

	// Base handles footer and padding calculation
	_, baseContentHeight := s.base.ContentSize()

	// Sidebar takes fixed width from the left
	contentWidth := s.width - SidebarWidth

	return contentWidth, baseContentHeight
}

// Render composes sidebar + content, then wraps in base layout
func (s *Sidebar) Render(content string) string {
	if s.width == 0 || s.height == 0 {
		return ""
	}

	// Get dimensions for content area (base handles padding/footer)
	baseContentWidth, baseContentHeight := s.base.ContentSize()

	// Calculate content width (remaining space after sidebar)
	contentWidth := baseContentWidth - SidebarWidth

	// Style content to fill remaining width and height
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(baseContentHeight)
	styledContent := contentStyle.Render(content)

	// Render sidebar
	sidebarView := s.sidebar.Render()

	// Compose sidebar (left) + content (right) using layers
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(sidebarView).X(0).Y(0),
		lipgloss.NewLayer(styledContent).X(SidebarWidth).Y(0),
	}

	canvas := lipgloss.NewCanvas(layers...)
	composedView := canvas.Render()

	// Wrap in base layout (adds footer and padding)
	return s.base.Render(composedView)
}
