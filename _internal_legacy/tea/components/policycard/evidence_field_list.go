package policycard

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
)

// viewFieldList renders per-field byte sizes for quality categories:
//
//	── Fields ───────────────────────────────────────────────────
//
//	3 fields, 4.2 KB/event (68% of event):
//
//	  http.request.body        3.8 KB
//	  http.response.headers    312 B
//	  http.user_agent          142 B
func (m *Model) viewFieldList(ev *domain.FieldListEvidence) string {
	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)
	sp := lipgloss.NewStyle().Background(m.theme.Bg)

	var lines []string

	// Header: "3 fields, 4.2 KB/event (68% of event):"
	header := fmt.Sprintf("%d fields, %s/event", len(ev.Fields), format.Bytes(ev.TotalBytes))
	if ev.BytesFraction > 0 {
		header += fmt.Sprintf(" (%s of event)", format.Percent(ev.BytesFraction))
	}
	header += ":"
	lines = append(lines, text.Render(header))
	lines = append(lines, "")

	// Field rows: key right-padded, byte size right-aligned.
	maxKey := 0
	maxVal := 0
	vals := make([]string, len(ev.Fields))
	for i, f := range ev.Fields {
		if len(f.Key) > maxKey {
			maxKey = len(f.Key)
		}
		vals[i] = format.Bytes(f.BytesPerEvent)
		if len(vals[i]) > maxVal {
			maxVal = len(vals[i])
		}
	}
	for i, f := range ev.Fields {
		lines = append(lines,
			sp.Render("  ")+
				accent.Render(padRight(f.Key, maxKey))+
				sp.Render("  ")+
				muted.Render(padRight(vals[i], maxVal)),
		)
	}

	body := strings.Join(lines, "\n")
	if body == "" {
		return ""
	}

	return renderDivider(m.theme, m.width, "Fields") + "\n\n" + body
}
