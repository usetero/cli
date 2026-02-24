package policycard

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
)

const maxConstantFields = 8

// viewConstantVaries renders the constant-vs-varies field breakdown:
//
//	── Example ──────────────────────────────────────────────────
//
//	15 of 22 fields identical across 10 examples:
//	  app                 broker-portal
//	  http.method         GET
//	  http.status_code    404
//	  ...and 12 more
//
//	7 vary: bid, date, duration, http.request_id, params, timestamp, xff
func (m *Model) viewConstantVaries(ev *domain.ConstantVariesEvidence) string {
	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	sp := lipgloss.NewStyle().Background(m.theme.Bg)

	total := len(ev.Constant) + len(ev.Varying)
	var lines []string

	// Constant fields — capped to keep the card scannable.
	if len(ev.Constant) > 0 {
		lines = append(lines, text.Render(fmt.Sprintf(
			"%d of %d fields identical across %d examples:",
			len(ev.Constant), total, ev.ExampleCount,
		)))
		lines = append(lines, "")

		shown := ev.Constant
		if len(shown) > maxConstantFields {
			shown = shown[:maxConstantFields]
		}

		maxKey := 0
		for _, f := range shown {
			if len(f.Key) > maxKey {
				maxKey = len(f.Key)
			}
		}
		for _, f := range shown {
			lines = append(lines, sp.Render("  ")+muted.Render(padRight(f.Key, maxKey))+sp.Render("  ")+text.Render(f.Value))
		}

		if remaining := len(ev.Constant) - maxConstantFields; remaining > 0 {
			lines = append(lines, sp.Render("  ")+muted.Render(fmt.Sprintf("...and %d more", remaining)))
		}
	}

	// Varying fields — one summary line with just the field names.
	if len(ev.Varying) > 0 {
		if len(ev.Constant) > 0 {
			lines = append(lines, "")
		}
		keys := make([]string, len(ev.Varying))
		for i, f := range ev.Varying {
			keys[i] = f.Key
		}
		lines = append(lines, text.Render(fmt.Sprintf("%d vary: ", len(ev.Varying)))+muted.Render(strings.Join(keys, ", ")))
	}

	body := strings.Join(lines, "\n")
	if body == "" {
		return ""
	}

	return m.renderExampleSection(body)
}
