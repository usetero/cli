package busy

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

func (m *Model) View() tea.View {
	if m.state == nil || m.width <= 0 {
		return tea.NewView("")
	}

	lines := []string{m.theme.Text.Section.Render(strings.TrimSpace(m.state.Label))}
	if status := strings.TrimSpace(m.state.Status); status != "" {
		lines = append(lines, "")
		lines = append(lines, present.Body(m.theme, status))
	}

	if pct, ok := m.percent(); ok {
		lines = append(lines, "")
		lines = append(lines, m.progress.ViewAs(pct))
		if detail := strings.TrimSpace(m.progressDetail()); detail != "" {
			lines = append(lines, "")
			lines = append(lines, m.theme.Text.Subtle.Render(detail))
		}
	}

	if detail := strings.TrimSpace(m.state.Detail); detail != "" {
		lines = append(lines, "")
		lines = append(lines, m.theme.Text.Muted.Render(detail))
	}

	return tea.NewView(present.Panel(m.theme, m.width, strings.Join(lines, "\n")))
}

func (m *Model) percent() (float64, bool) {
	if m.state == nil || m.state.Progress == nil || m.state.Progress.Total <= 0 {
		return 0, false
	}
	return float64(m.state.Progress.Current) / float64(m.state.Progress.Total) * 100, true
}

func (m *Model) progressDetail() string {
	if m.state == nil || m.state.Progress == nil || m.state.Progress.Total <= 0 {
		return ""
	}
	return fmt.Sprintf("%d / %d rows", m.state.Progress.Current, m.state.Progress.Total)
}
