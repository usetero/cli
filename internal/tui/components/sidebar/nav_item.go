package sidebar

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

// NavItem represents a single navigation item in the sidebar
type NavItem struct {
	label     string
	stat      string      // Optional stat to display on the right (e.g., "2", "1.54m/hr", "23% ↑2%")
	statColor color.Color // Color for the stat (e.g., theme.Error.Fg for red, nil for default muted)
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

	// Shortcut (if present) - muted text on panel
	shortcutStyle := lipgloss.NewStyle().Foreground(theme.Panel.TextMuted)
	if n.shortcut.Keys() != nil && len(n.shortcut.Keys()) > 0 {
		// Get the help text (e.g., "⌥1")
		shortcutText := n.shortcut.Help().Key
		leftSide = shortcutStyle.Render(shortcutText) + " "
	}

	// Active items use accent color, inactive use panel text
	var labelStyle, statStyle lipgloss.Style
	if n.active {
		labelStyle = lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
		statStyle = lipgloss.NewStyle().Foreground(theme.Accent)
	} else {
		labelStyle = lipgloss.NewStyle().Foreground(theme.Panel.Text)
		// Use custom stat color if provided, otherwise default to muted
		if n.statColor != nil {
			statStyle = lipgloss.NewStyle().Foreground(n.statColor)
		} else {
			statStyle = lipgloss.NewStyle().Foreground(theme.Panel.TextMuted)
		}
	}

	// Add label
	leftSide += labelStyle.Render(n.label)

	// Add indicator dot if needed
	if n.indicator {
		indicatorStyle := lipgloss.NewStyle().Foreground(theme.Error.Fg)
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
