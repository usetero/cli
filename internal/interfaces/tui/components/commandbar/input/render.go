package input

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/cursor"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

const inputPadX = 1

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
	m.textarea.SetWidth(present.PanelInnerWidth(width))
}

// View renders the command input block.
func (m *Model) View() tea.View {
	if m.width <= 0 {
		return tea.NewView("")
	}

	var content string
	if m.textarea.Value() == "" {
		content = lipgloss.NewStyle().
			Foreground(m.theme.Palette.TextMuted).
			Background(m.theme.Background).
			Render(m.placeholder)
	} else {
		if m.secret {
			content = lipgloss.NewStyle().
				Foreground(m.theme.Palette.Text).
				Background(m.theme.Background).
				Render(strings.Repeat("•", len([]rune(m.textarea.Value()))))
		} else {
			content = m.textarea.View()
		}

		r, g, b, _ := m.theme.Background.RGBA()
		bgSeq := fmt.Sprintf("\033[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
		content = strings.ReplaceAll(content, "\033[0m", "\033[0m"+bgSeq)
	}

	if cur := m.textarea.Cursor(); cur != nil {
		content = cursor.Insert(content, cur.X, cur.Y)
	}

	lines := strings.Split(content, "\n")
	visibleLines := m.visibleLines()
	for len(lines) < visibleLines {
		lines = append(lines, " ")
	}
	content = strings.Join(lines[:visibleLines], "\n")

	return tea.NewView(
		lipgloss.NewStyle().
			Width(present.PanelInnerWidth(m.width)).
			Padding(0, inputPadX).
			Background(m.theme.Background).
			Foreground(m.theme.Palette.Text).
			Render(content),
	)
}
