package layouts

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/header"
)

// Header is a layout with a header at the top and content below.
// Uses Base to add footer+padding to everything.
type Header struct {
	// Nested layouts
	base *Base

	// Child components
	header *header.Component

	// State
	width  int
	height int
}

// NewHeader creates a new header layout
func NewHeader(logger log.Logger) *Header {
	return &Header{
		base:   NewBase(logger),
		header: header.New(logger),
	}
}

// Update handles messages for the header layout
func (h *Header) Update(msg tea.Msg) tea.Cmd {
	// Cascade to content area
	return h.base.Update(msg)
}

// SetSize sets the dimensions for the layout
func (h *Header) SetSize(width, height int) {
	h.width = width
	h.height = height

	// Base gets full dimensions
	h.base.SetSize(width, height)

	// Header gets width after base padding (calculate directly to avoid calling ContentSize during init)
	baseWidth := width - (horizontalPadding * 2)
	h.header.Update(tea.WindowSizeMsg{Width: baseWidth, Height: height})
}

// SetKeyBindings updates the key bindings shown in the footer
func (h *Header) SetKeyBindings(bindings []key.Binding) {
	h.base.SetKeyBindings(bindings)
}

// SetError sets the error to display in the footer
func (h *Header) SetError(err error) {
	h.base.SetError(err)
}

// ContentSize returns the available space for content (width x height after header and footer)
func (h *Header) ContentSize() (int, int) {
	if h.width == 0 || h.height == 0 {
		return 0, 0
	}

	// Base handles footer and padding calculation
	baseWidth, baseHeight := h.base.ContentSize()

	// Header takes some height from the top
	headerView := h.header.View()
	headerHeight := lipgloss.Height(headerView)

	contentHeight := baseHeight - headerHeight

	return baseWidth, contentHeight
}

// Render composes header + content, then wraps in base layout
func (h *Header) Render(content string) string {
	if h.width == 0 || h.height == 0 {
		return ""
	}

	// Render header
	headerView := h.header.View()

	// Compose header (top) + content (below)
	composedView := lipgloss.JoinVertical(
		lipgloss.Left,
		headerView,
		content,
	)

	// Wrap in base layout (adds footer and padding)
	return h.base.Render(composedView)
}
