package commandbar

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

const commandbarRailGlyph = "┃"
const (
	commandbarPadX      = 1
	commandbarPadTop    = 1
	commandbarPadBottom = 1
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
	surfaceWidth := m.surfaceWidth()
	if m.visorEnabled {
		m.children.visor.SetSize(surfaceWidth, m.children.visor.PreferredHeight(surfaceWidth))
	}
	m.children.action.SetSize(surfaceWidth, m.interactionHeight(surfaceWidth))
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
	surfaceWidth := surfaceWidthFor(width)
	height := 0
	if m.visorEnabled {
		if visorHeight := m.children.visor.PreferredHeight(surfaceWidth); visorHeight > 0 {
			height += visorHeight
		}
	}
	if noticeHeight := m.noticeHeight(); noticeHeight > 0 {
		if height > 0 {
			height++
		}
		height += noticeHeight
	}
	if interactiveHeight := m.interactionHeight(surfaceWidth); interactiveHeight > 0 {
		if height > 0 {
			height++
		}
		height += interactiveHeight
	}
	height += commandbarPadTop + commandbarPadBottom
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) renderContent() string {
	content := m.renderStack()
	content = bottomAnchor(content, m.surfaceWidth(), innerCommandbarHeight(m.height))
	return lipgloss.NewStyle().
		Width(m.width).
		Background(m.surfaceTheme.Background).
		Padding(commandbarPadTop, commandbarPadX, commandbarPadBottom, commandbarPadX).
		Render(content)
}

func (m *Model) renderStack() string {
	parts := make([]string, 0, 3)
	hasVisor := false

	if m.visorEnabled {
		if visor := m.children.visor.View().Content; strings.TrimSpace(visor) != "" {
			parts = append(parts, visor)
			hasVisor = true
		}
	}

	if notice := m.renderNotice(); strings.TrimSpace(notice) != "" {
		if hasVisor {
			parts = append(parts, "")
		}
		parts = append(parts, notice)
		hasVisor = true
	}

	if m.err != nil {
		if hasVisor {
			parts = append(parts, "")
		}
		parts = append(parts, m.renderInteractive(m.children.action.err.View().Content))
	} else if m.busy != nil {
		if hasVisor {
			parts = append(parts, "")
		}
		parts = append(parts, m.renderInteractive(m.children.action.busy.View().Content))
	} else {
		switch m.mode {
		case ModeError:
			if hasVisor {
				parts = append(parts, "")
			}
			parts = append(parts, m.renderInteractive(m.children.action.err.View().Content))
		case ModeAction:
			if hasVisor {
				parts = append(parts, "")
			}
			parts = append(parts, m.renderAction())
		case ModeInput:
			if hasVisor {
				parts = append(parts, "")
			}
			parts = append(parts, m.renderInteractive(m.children.action.input.View().Content))
		case ModeSelect:
			if hasVisor {
				parts = append(parts, "")
			}
			if child := m.active(); child != nil {
				parts = append(parts, m.renderInteractive(m.renderSelect(child.View().Content)))
			}
		}
	}

	if len(parts) == 0 {
		return " "
	}
	return lipgloss.NewStyle().
		Width(m.surfaceWidth()).
		Background(m.surfaceTheme.Background).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m *Model) renderAction() string {
	if m.input == nil || strings.TrimSpace(m.input.Action) == "" {
		return ""
	}
	keycap := lipgloss.NewStyle().
		Foreground(m.actionColor()).
		Background(m.surfaceTheme.Background).
		Bold(true).
		Render("[enter]")

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		keycap,
		" ",
		lipgloss.NewStyle().
			Foreground(m.actionColor()).
			Background(m.surfaceTheme.Background).
			Render(strings.TrimSpace(m.input.Action)),
	)

	return m.renderInteractive(
		lipgloss.NewStyle().
			Width(m.surfaceWidth()).
			Background(m.surfaceTheme.Background).
			Foreground(m.surfaceTheme.Palette.Text).
			Render(content),
	)
}

func (m *Model) renderSelect(content string) string {
	return present.Panel(
		m.surfaceTheme,
		m.surfaceWidth(),
		strings.TrimRight(content, "\n"),
	)
}

// surfaceWidth returns the width available inside the command bar's outer left
// border. Child surfaces render against this width; they should not subtract
// for the border again.
func (m *Model) surfaceWidth() int {
	return surfaceWidthFor(m.width)
}

func surfaceWidthFor(width int) int {
	width = width - lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: commandbarRailGlyph}).
		GetHorizontalFrameSize()
	width = width - commandbarPadX*2
	if width < 1 {
		return 1
	}
	return width
}

func innerCommandbarHeight(height int) int {
	height = height - commandbarPadTop - commandbarPadBottom
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) noticeHeight() int {
	notice := m.notice
	if m.localNotice != nil {
		notice = m.localNotice
	}
	if notice == nil || strings.TrimSpace(notice.Message) == "" {
		return 0
	}
	return 1
}

func (m *Model) interactionHeight(width int) int {
	if m.err != nil {
		return m.children.action.err.PreferredHeight(width)
	}
	if m.busy != nil {
		return m.children.action.busy.PreferredHeight(width)
	}

	switch m.mode {
	case ModeError:
		return m.children.action.err.PreferredHeight(width)
	case ModeAction:
		return 3
	case ModeInput:
		return m.children.action.input.PreferredHeight(width)
	case ModeSelect:
		if child, ok := m.active().(core.HeightProvider); ok {
			return child.PreferredHeight(width) + 2
		}
	}
	return 0
}

func (m *Model) renderInteractive(content string) string {
	return m.renderInteractiveWithRail(content, m.railColor())
}

func (m *Model) renderInteractiveWithRail(content string, railColor color.Color) string {
	block := lipgloss.NewStyle().
		Width(m.surfaceWidth()).
		Background(m.surfaceTheme.Background).
		Render(content)

	return lipgloss.NewStyle().
		Background(m.surfaceTheme.Background).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: commandbarRailGlyph}).
		BorderForeground(railColor).
		BorderBackground(m.surfaceTheme.Background).
		Render(block)
}

func (m *Model) railColor() color.Color {
	switch m.surfaceState() {
	case SurfaceError:
		return m.surfaceTheme.Palette.Error
	case SurfaceBusy:
		return m.surfaceTheme.Palette.TextMuted
	case SurfaceActive:
		return m.surfaceTheme.Palette.Brand
	default:
		return m.surfaceTheme.Palette.Border
	}
}

func (m *Model) actionColor() color.Color {
	switch m.surfaceState() {
	case SurfaceError:
		return m.surfaceTheme.Palette.Error
	default:
		return m.surfaceTheme.Palette.Brand
	}
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
