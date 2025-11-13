package layouts

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/footer"
	"github.com/usetero/cli/internal/tui/keymap"
)

const (
	// Global padding applied by Base layout
	horizontalPadding = 2
	verticalPadding   = 1
)

// Base is the foundation layout that wraps content with a footer and padding.
// All other layouts compose with Base to get consistent footer and padding.
type Base struct {
	// Child components
	footer *footer.Component

	// State
	width  int
	height int
}

// NewBase creates a new base layout
func NewBase(logger log.Logger) *Base {
	return &Base{
		footer: footer.New(logger),
	}
}

// Update handles messages for the content area layout
func (c *Base) Update(msg tea.Msg) tea.Cmd {
	return c.footer.Update(msg)
}

// SetSize sets the dimensions for the layout
func (c *Base) SetSize(width, height int) {
	c.width = width
	c.height = height
	// Pass inner width to footer (after accounting for horizontal padding)
	innerWidth := width - (horizontalPadding * 2)
	c.footer.Update(tea.WindowSizeMsg{Width: innerWidth, Height: height})
}

// SetKeyBindings updates the key bindings shown in the footer
func (c *Base) SetKeyBindings(bindings []key.Binding) {
	keyMap := keymap.Simple{Keys: bindings}
	c.footer.SetKeyMap(keyMap)
}

// SetError sets the error to display in the footer
func (c *Base) SetError(err error) {
	c.footer.SetError(err)
}

// ContentSize returns the available space for content (width x height after padding and footer)
func (c *Base) ContentSize() (int, int) {
	if c.width == 0 || c.height == 0 {
		return 0, 0
	}

	// Account for global padding
	contentWidth := c.width - (horizontalPadding * 2)

	// Calculate actual footer height
	footerView := c.footer.View()
	footerHeight := lipgloss.Height(footerView)
	footerSpacing := 1 // One blank line above footer

	// Account for vertical padding (top and bottom) and footer
	contentHeight := c.height - (verticalPadding * 2) - footerHeight - footerSpacing
	return contentWidth, contentHeight
}

// Render wraps content with a footer and applies global padding
func (c *Base) Render(content string) string {
	if c.width == 0 || c.height == 0 {
		return ""
	}

	// Calculate dimensions inside padding
	innerWidth := c.width - (horizontalPadding * 2)

	// Render footer
	footerView := c.footer.View()
	footerHeight := lipgloss.Height(footerView)
	footerSpacing := 1 // One blank line above footer

	// Calculate content height (inside padding, minus footer)
	contentHeight := c.height - (verticalPadding * 2) - footerHeight - footerSpacing

	// Ensure content fills available height
	contentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(contentHeight)
	styledContent := contentStyle.Render(content)

	// Compose content + spacing + footer
	components := []string{
		styledContent,
		"", // Spacing above footer
		footerView,
	}

	innerView := lipgloss.JoinVertical(lipgloss.Top, components...)

	// Apply global padding to entire view
	return lipgloss.NewStyle().
		Padding(verticalPadding, horizontalPadding, verticalPadding, horizontalPadding).
		Render(innerView)
}
