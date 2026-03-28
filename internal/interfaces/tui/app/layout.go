package app

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

func (m *Model) SetSize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	m.width = width
	m.height = height
	m.applyLayout()
}

func (m *Model) applyLayout() {
	if m.width <= 0 {
		return
	}

	contentWidth := m.innerWidth()
	m.children.statusbar.SetSize(contentWidth, 1)

	helpHeight := m.helpbarHeight(contentWidth)
	commandHeight := m.commandbarHeight(contentWidth, helpHeight)
	m.children.commandbar.SetSize(contentWidth, commandHeight)
	m.children.helpbar.SetSize(contentWidth, helpHeight)
	m.body.SetSize(contentWidth, m.bodyHeight(commandHeight, helpHeight))
}

func (m *Model) innerWidth() int {
	width := m.width - m.theme.Shell.Outer.GetHorizontalFrameSize()
	if width < 0 {
		return 0
	}
	return width
}

func (m *Model) innerHeight() int {
	height := m.height - m.theme.Shell.Outer.GetVerticalFrameSize()
	if height < 0 {
		return 0
	}
	return height
}

func (m *Model) headerHeight(width int) int {
	height := lipgloss.Height(
		m.theme.Shell.HeaderBar.Width(width).Render(normalizeLine(m.children.statusbar.View().Content)),
	)
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) helpbarHeight(width int) int {
	m.children.helpbar.SetSize(width, 0)
	height := lipgloss.Height(strings.TrimRight(m.children.helpbar.View().Content, "\n"))
	if height < 0 {
		return 0
	}
	return height
}

func (m *Model) commandbarHeight(width, helpHeight int) int {
	provider, ok := any(m.children.commandbar).(core.HeightProvider)
	if !ok {
		return 1
	}
	preferred := provider.PreferredHeight(width)
	maxHeight := m.innerHeight() - m.headerHeight(width) - m.footerChromeHeight(helpHeight) - helpHeight - m.footerSpacerHeight(helpHeight) - 1
	if maxHeight < 1 {
		maxHeight = 1
	}
	if preferred < 1 {
		return 1
	}
	if preferred > maxHeight {
		return maxHeight
	}
	return preferred
}

func (m *Model) footerSpacerHeight(helpHeight int) int {
	if helpHeight > 0 {
		return 1
	}
	return 0
}

func (m *Model) bodyHeight(commandHeight, helpHeight int) int {
	height := m.innerHeight() - m.headerHeight(m.innerWidth()) - m.footerChromeHeight(helpHeight) - commandHeight - helpHeight - m.footerSpacerHeight(helpHeight)
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) footerChromeHeight(helpHeight int) int {
	return 2
}

func normalizeLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return " "
	}
	return value
}

func normalizeBody(value string) string {
	value = strings.TrimRight(value, "\n")
	if value == "" {
		return " "
	}
	return value
}
