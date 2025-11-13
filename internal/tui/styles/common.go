package styles

import "github.com/charmbracelet/lipgloss/v2"

// CommonStyles provides commonly used text styles for the TUI.
// These styles are cached and reused across the application for consistency and performance.
type CommonStyles struct {
	Title    lipgloss.Style // Page/step titles (Primary + Bold)
	Subtitle lipgloss.Style // Section descriptions (TextSubtle)
	Body     lipgloss.Style // Main text content (Text)
	Action   lipgloss.Style // User action prompts like "Press Enter..." (Primary)
	Help     lipgloss.Style // Secondary help text (TextMuted)
	URL      lipgloss.Style // URL displays (TextSubtle)
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
			Title:    lipgloss.NewStyle().Foreground(theme.Primary).Bold(true),
			Subtitle: lipgloss.NewStyle().Foreground(theme.TextSubtle),
			Body:     lipgloss.NewStyle().Foreground(theme.Text),
			Action:   lipgloss.NewStyle().Foreground(theme.Primary),
			Help:     lipgloss.NewStyle().Foreground(theme.TextMuted),
			URL:      lipgloss.NewStyle().Foreground(theme.TextSubtle),
			Success:  lipgloss.NewStyle().Foreground(theme.Success).Bold(true),
			Error:    lipgloss.NewStyle().Foreground(theme.Error),
		}
	}
	return commonStyles
}
