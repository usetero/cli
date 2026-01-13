// Package styles provides consistent styling for CLI output.
// Used by both TUI components and traditional CLI commands.
package styles

import "charm.land/lipgloss/v2"

// CommonStyles provides commonly used text styles for the TUI.
// These styles are cached and reused across the application for consistency and performance.
type CommonStyles struct {
	Title    lipgloss.Style // Page/step titles (Accent + Bold)
	Subtitle lipgloss.Style // Section descriptions (TextMuted)
	Body     lipgloss.Style // Main text content (Text)
	Action   lipgloss.Style // User action prompts like "Press Enter..." (Accent)
	Help     lipgloss.Style // Secondary help text (TextMuted)
	URL      lipgloss.Style // URL displays (TextMuted)
	Success  lipgloss.Style // Success messages with checkmarks (Success + Bold)
	Error    lipgloss.Style // Error messages (Error)
}

var commonStyles *CommonStyles

// Common returns the cached common styles.
// Styles are created once on first call and reused for performance.
func Common() *CommonStyles {
	if commonStyles == nil {
		theme := CurrentTheme()
		commonStyles = &CommonStyles{
			Title:    lipgloss.NewStyle().Foreground(theme.Accent).Bold(true),
			Subtitle: lipgloss.NewStyle().Foreground(theme.Page.TextMuted),
			Body:     lipgloss.NewStyle().Foreground(theme.Page.Text),
			Action:   lipgloss.NewStyle().Foreground(theme.Accent),
			Help:     lipgloss.NewStyle().Foreground(theme.Page.TextMuted),
			URL:      lipgloss.NewStyle().Foreground(theme.Page.TextMuted),
			Success:  lipgloss.NewStyle().Foreground(theme.Success.Fg).Bold(true),
			Error:    lipgloss.NewStyle().Foreground(theme.Error.Fg),
		}
	}
	return commonStyles
}
