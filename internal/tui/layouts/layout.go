package layouts

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// Layout defines the interface that all layouts must implement.
// Layouts compose header/sidebar/footer around content and manage spacing.
type Layout interface {
	// Update handles messages
	Update(tea.Msg) tea.Cmd

	// SetSize sets the terminal dimensions
	SetSize(width, height int)

	// SetKeyBindings updates the key bindings shown in the footer
	SetKeyBindings(bindings []key.Binding)

	// SetError sets the error to display in the footer
	SetError(err error)

	// ContentSize returns the available space for content (width, height)
	ContentSize() (int, int)

	// Render wraps content with the layout (header, footer, padding, etc.)
	Render(content string) string
}
