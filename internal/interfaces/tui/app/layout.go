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
	m.surface.shell.statusbar.SetSize(contentWidth, 1)
	m.surface.shell.divider.SetSize(contentWidth, 1)
	title := windowTitle
	if page := m.surface.body.Page(); page.Title != "" {
		title = m.surface.shell.commandbar.FooterTitle(page.Title)
	}
	m.surface.shell.divider.SetState(title, m.surface.body.Busy() != nil, m.surface.shell.commandbar.IsPaletteOpen(), m.titleSpinnerStep)

	helpHeight := m.helpbarHeight(contentWidth)
	commandHeight := m.commandbarHeight(contentWidth, helpHeight)
	m.surface.shell.commandbar.SetSize(contentWidth, commandHeight)
	m.surface.shell.helpbar.SetSize(contentWidth, helpHeight)
	bodyHeight := m.bodyHeight(commandHeight, helpHeight)
	m.surface.shell.divider.SetSize(contentWidth, 1)
	m.surface.body.SetSize(m.bodyContentWidth(contentWidth), m.bodyContentViewportHeight(bodyHeight))
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
		m.theme.Shell.HeaderBar.Width(width).Render(normalizeLine(m.surface.shell.statusbar.View().Content)),
	)
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) helpbarHeight(width int) int {
	m.surface.shell.helpbar.SetSize(width, 0)
	height := lipgloss.Height(strings.TrimRight(m.surface.shell.helpbar.View().Content, "\n"))
	if height < 0 {
		return 0
	}
	return height
}

func (m *Model) commandbarHeight(width, helpHeight int) int {
	provider, ok := any(m.surface.shell.commandbar).(core.HeightProvider)
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
	return 0
}

func (m *Model) bodyContentWidth(width int) int {
	width = width - m.theme.Shell.Body.GetHorizontalFrameSize() - bodyPadX*2
	if width < 1 {
		return 1
	}
	return width
}

func (m *Model) bodyContentHeight(height int) int {
	height = height - m.theme.Shell.Body.GetVerticalFrameSize() - bodyPadY*2
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) bodyContentViewportHeight(height int) int {
	height = m.bodyContentHeight(height) - m.bodyDividerHeight() - m.bodyDividerGapHeight()
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) bodyDividerHeight() int {
	return 1
}

func (m *Model) bodyDividerGapHeight() int {
	return 1
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
