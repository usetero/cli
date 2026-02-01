package chat

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

// Compile-time check that Divider implements Item.
var _ Item = (*Divider)(nil)

// Divider is a visual separator between conversation turns.
// Shows optional metadata like model name and response time.
type Divider struct {
	theme *styles.Theme
	width int

	id       string
	model    string
	duration time.Duration
}

// NewDivider creates a new divider component.
func NewDivider(theme *styles.Theme, id string) *Divider {
	return &Divider{
		theme: theme,
		id:    id,
	}
}

// ID returns the divider ID.
func (d *Divider) ID() string {
	return d.id
}

// Init initializes the component (no-op for divider).
func (d *Divider) Init() tea.Cmd {
	return nil
}

// Update handles messages (no-op for divider).
func (d *Divider) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the divider.
func (d *Divider) View() string {
	colors := d.theme.Colors

	// Build metadata parts
	var parts []string

	if d.model != "" {
		parts = append(parts, d.model)
	}

	if d.duration > 0 {
		parts = append(parts, formatDuration(d.duration))
	}

	// If no metadata, render simple line
	if len(parts) == 0 {
		return lipgloss.NewStyle().
			Foreground(colors.BorderDefault).
			Width(d.width).
			Render("─")
	}

	// Render metadata with subtle styling
	metadata := ""
	for i, part := range parts {
		if i > 0 {
			metadata += " · "
		}
		metadata += part
	}

	return lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Width(d.width).
		Align(lipgloss.Center).
		Render(metadata)
}

// SetWidth sets the available width for rendering.
func (d *Divider) SetWidth(width int) {
	d.width = width
}

// SetModel sets the model name to display.
func (d *Divider) SetModel(model string) {
	d.model = model
}

// SetDuration sets the response duration to display.
func (d *Divider) SetDuration(duration time.Duration) {
	d.duration = duration
}

// Spinning returns false (dividers don't animate).
func (d *Divider) Spinning() bool {
	return false
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
