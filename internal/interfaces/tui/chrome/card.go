package chrome

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// RenderCard returns a standard card with optional title and body text.
func RenderCard(t theme.Theme, title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if body == "" {
		body = " "
	}

	lines := make([]string, 0, 3)
	if title != "" {
		lines = append(lines, t.Card.Title.Render(title), "")
	}
	lines = append(lines, t.Card.Body.Render(body))
	return t.Card.Container.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// RenderErrorCard returns an error-toned card with optional title and body text.
func RenderErrorCard(t theme.Theme, title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if body == "" {
		body = " "
	}

	lines := make([]string, 0, 3)
	if title != "" {
		lines = append(lines, t.Card.ErrorTitle.Render(title), "")
	}
	lines = append(lines, t.Card.Body.Render(body))
	return t.Card.ErrorContainer.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
