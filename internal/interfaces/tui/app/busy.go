package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var windowTitleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var footerLeadFrames = []string{"◇", "◈", "◆", "◈"}

type windowTitleTickMsg struct{}

func (m *Model) startWindowTitleSpinner() tea.Cmd {
	if m.surface.body.Busy() == nil {
		m.titleSpinnerTick = false
		m.titleSpinnerStep = 0
		return nil
	}
	if m.titleSpinnerTick {
		return nil
	}
	m.titleSpinnerTick = true
	return m.windowTitleTick()
}

func (m *Model) stopWindowTitleSpinner() {
	m.titleSpinnerTick = false
	m.titleSpinnerStep = 0
}

func (m *Model) updateWindowTitleSpinner(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(windowTitleTickMsg); !ok {
		return nil
	}
	if m.surface.body.Busy() == nil {
		m.stopWindowTitleSpinner()
		return nil
	}
	m.titleSpinnerStep = (m.titleSpinnerStep + 1) % len(windowTitleFrames)
	m.titleSpinnerTick = true
	return m.windowTitleTick()
}

func (m *Model) windowTitleTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return windowTitleTickMsg{}
	})
}

func (m *Model) windowTitle() string {
	if m.surface.body.Busy() == nil {
		return windowTitle
	}
	return windowTitleFrames[m.titleSpinnerStep] + " " + windowTitle
}

func (m *Model) footerRuleLabel() string {
	lead := footerLeadFrames[0]
	leadStyle := m.theme.Shell.FooterLead
	if m.surface.body.Busy() != nil {
		lead = footerLeadFrames[m.titleSpinnerStep%len(footerLeadFrames)]
		leadStyle = lipgloss.NewStyle().
			Foreground(m.theme.Palette.Brand).
			Background(m.theme.Background).
			Bold(true)
	}
	if m.surface.shell.commandbar.IsPaletteOpen() {
		lead = "‹"
		leadStyle = m.theme.Shell.FooterLead
	}

	title := windowTitle
	if page := m.surface.body.Page(); page.Title != "" {
		title = page.Title
	}

	return leadStyle.Render(lead) + " " + m.theme.Shell.FooterLead.Render(m.surface.shell.commandbar.FooterTitle(title))
}
