package error

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

func (m *Model) View() tea.View {
	if m.state == nil || m.width <= 0 {
		return tea.NewView("")
	}

	lines := []string{
		present.Error(m.theme, "Error"),
	}
	if message := strings.TrimSpace(m.state.Message); message != "" {
		lines = append(lines, "")
		lines = append(lines, present.Body(m.theme, message))
	}
	if detail := strings.TrimSpace(m.state.Detail); detail != "" {
		lines = append(lines, "")
		lines = append(lines, m.theme.Text.Muted.Render(detail))
	}
	if action := strings.TrimSpace(m.state.Action); action != "" {
		lines = append(lines, "")
		lines = append(lines, present.Body(m.theme,
			m.theme.Text.Error.Render("[enter]")+" "+m.theme.Text.Error.Render(action),
		))
	}

	return tea.NewView(present.Panel(m.theme, m.width, strings.Join(lines, "\n")))
}
