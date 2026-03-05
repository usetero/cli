package policycard

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
)

// viewHighlightedExample renders a single log example with relevant fields
// visually highlighted. Relevant fields appear first in accent color,
// then remaining fields in muted.
//
//	── Example ──────────────────────────────────────────────────
//
//	  app                 broker-portal
//	  http.method         GET        (accent if relevant)
//	  http.status_code    404
func (m *Model) viewHighlightedExample(ev *domain.HighlightedExampleEvidence) string {
	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)
	sp := lipgloss.NewStyle().Background(m.theme.Bg)

	relevant := make(map[string]bool, len(ev.RelevantKeys))
	for _, k := range ev.RelevantKeys {
		relevant[k.Key()] = true
	}

	maxKey := 0
	for _, f := range ev.Attrs {
		if len(f.Key) > maxKey {
			maxKey = len(f.Key)
		}
	}

	var lines []string
	for _, f := range ev.Attrs {
		keyStyle := muted
		valStyle := text
		if relevant[f.Key] {
			keyStyle = accent
			valStyle = accent
		}
		lines = append(lines, sp.Render("  ")+keyStyle.Render(padRight(f.Key, maxKey))+sp.Render("  ")+valStyle.Render(f.Value))
	}

	body := strings.Join(lines, "\n")
	if body == "" {
		return ""
	}

	return m.renderExampleSection(body)
}
