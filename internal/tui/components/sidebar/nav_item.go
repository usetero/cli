package sidebar

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/tui/styles"
)

// NavItem represents a single navigation item in the sidebar
type NavItem struct {
	label     string
	stat      string      // Optional stat to display on the right (e.g., "2", "1.54m/hr", "23% ↑2%")
	statColor color.Color // Color for the stat (theme.Error for red, nil for default theme.Field)
	active    bool
	indicator bool        // If true, shows a red dot (e.g., for unread messages or new activity)
	shortcut  key.Binding // Optional keyboard shortcut (e.g., Alt+1)
}

// NewNavItem creates a new navigation item
func NewNavItem(label string, stat string, statColor color.Color, active bool, indicator bool, shortcut key.Binding) NavItem {
	return NavItem{
		label:     label,
		stat:      stat,
		statColor: statColor,
		active:    active,
		indicator: indicator,
		shortcut:  shortcut,
	}
}

// Render renders the navigation item with left-aligned label and right-aligned stat
// Format: "⌥1 Chat                2" or "⌥1 Chat•               2"
func (n NavItem) Render(width int, theme *styles.Theme) string {
	// Build the left side: shortcut + label + indicator
	var leftSide string

	// Shortcut (if present)
	shortcutStyle := lipgloss.NewStyle().Foreground(theme.Field)
	if n.shortcut.Keys() != nil && len(n.shortcut.Keys()) > 0 {
		// Get the help text (e.g., "⌥1")
		shortcutText := n.shortcut.Help().Key
		leftSide = shortcutStyle.Render(shortcutText) + " "
	}

	// Active items use primary color, inactive use text color
	var labelStyle, statStyle lipgloss.Style
	if n.active {
		labelStyle = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
		statStyle = lipgloss.NewStyle().Foreground(theme.Primary)
	} else {
		labelStyle = lipgloss.NewStyle().Foreground(theme.Text)
		// Use custom stat color if provided, otherwise default to theme.Field
		if n.statColor != nil {
			statStyle = lipgloss.NewStyle().Foreground(n.statColor)
		} else {
			statStyle = lipgloss.NewStyle().Foreground(theme.Field)
		}
	}

	// Add label
	leftSide += labelStyle.Render(n.label)

	// Add indicator dot if needed
	if n.indicator {
		indicatorStyle := lipgloss.NewStyle().Foreground(theme.Error)
		leftSide += indicatorStyle.Render("•")
	}

	// Calculate spacing
	leftWidth := lipgloss.Width(leftSide)
	statWidth := lipgloss.Width(n.stat)
	spacingWidth := width - leftWidth - statWidth
	if spacingWidth < 0 {
		spacingWidth = 0
	}

	// If there's no stat, just return the left side
	if n.stat == "" {
		return leftSide
	}

	return leftSide + strings.Repeat(" ", spacingWidth) + statStyle.Render(n.stat)
}
