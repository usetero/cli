package commandbar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	m.width = width
	m.height = height
	m.children.visor.SetSize(width, height)
	m.children.input.SetSize(m.interactiveWidth(), height)
	m.children.selectlist.SetSize(m.interactiveWidth(), height)
}

// View renders the current command bar surface.
func (m *Model) View() tea.View {
	if m.width <= 0 {
		return tea.NewView("")
	}
	return tea.NewView(m.renderContent())
}

// PreferredHeight reports the rendered footer height for the given width.
func (m *Model) PreferredHeight(width int) int {
	if width < 1 {
		return 1
	}

	prevWidth, prevHeight := m.width, m.height
	m.SetSize(width, 0)
	height := lipgloss.Height(m.renderStack())
	m.SetSize(prevWidth, prevHeight)
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) renderContent() string {
	content := m.renderStack()
	content = bottomAnchor(content, m.width, m.height)
	return lipgloss.NewStyle().
		Width(m.width).
		Render(content)
}

func (m *Model) renderStack() string {
	parts := make([]string, 0, 3)
	hasVisor := false

	if visor := m.children.visor.View().Content; strings.TrimSpace(visor) != "" {
		parts = append(parts, visor)
		hasVisor = true
	}

	if m.busy != nil {
		if hasVisor {
			parts = append(parts, "")
		}
		parts = append(parts, m.renderBusy())
	} else {
		switch m.mode {
		case ModeAction:
			if hasVisor {
				parts = append(parts, "")
			}
			parts = append(parts, m.renderAction())
		case ModeInput:
			if hasVisor {
				parts = append(parts, "")
			}
			parts = append(parts, m.renderInteractive(m.children.input.View().Content))
		case ModeSelect:
			if hasVisor {
				parts = append(parts, "")
			}
			parts = append(parts, m.renderInteractive(m.children.selectlist.View().Content))
		}
	}

	if len(parts) == 0 {
		return " "
	}
	return lipgloss.NewStyle().
		Width(m.width).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m *Model) renderAction() string {
	if m.input == nil || strings.TrimSpace(m.input.Action) == "" {
		return ""
	}
	return present.Panel(
		m.surfaceTheme,
		m.width,
		m.surfaceTheme.Text.Section.Render(strings.TrimSpace(m.input.Action)),
	)
}

func (m *Model) renderBusy() string {
	if m.busy == nil {
		return ""
	}

	lines := []string{m.surfaceTheme.Text.Section.Render(m.busy.Label)}
	if strings.TrimSpace(m.busy.Detail) != "" {
		lines = append(lines, m.surfaceTheme.Text.Body.Render(m.busy.Detail))
	}

	return m.renderInteractive(present.Panel(m.surfaceTheme, m.interactiveWidth(), lipgloss.JoinVertical(lipgloss.Left, lines...)))
}

func (m *Model) interactiveWidth() int {
	width := m.width - lipgloss.NewStyle().BorderLeft(true).GetHorizontalFrameSize()
	if width < 1 {
		return 1
	}
	return width
}

func (m *Model) renderInteractive(content string) string {
	block := lipgloss.NewStyle().
		Width(m.interactiveWidth()).
		Background(m.surfaceTheme.Background).
		Render(content)

	return lipgloss.NewStyle().
		Width(m.width).
		Background(m.theme.Background).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.surfaceTheme.Palette.Accent).
		BorderBackground(m.theme.Background).
		Render(block)
}

func bottomAnchor(content string, width, height int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if height > 0 && len(lines) > height {
		lines = lines[len(lines)-height:]
	}
	content = lipgloss.JoinVertical(lipgloss.Left, lines...)

	if height > 0 {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			AlignHorizontal(lipgloss.Left).
			AlignVertical(lipgloss.Bottom).
			Render(content)
	}

	return lipgloss.NewStyle().
		Width(width).
		Render(content)
}
