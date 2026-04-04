package commandbar

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

const noticeGutterWidth = 2

func (m *Model) SetLocalNotice(notice *core.Notice) {
	m.localNotice = notice
}

func (m *Model) renderNotice() string {
	notice := m.notice
	if m.localNotice != nil {
		notice = m.localNotice
	}
	if notice == nil || strings.TrimSpace(notice.Message) == "" {
		return ""
	}

	style := m.theme.Text.Body
	glyphStyle := m.theme.Text.Muted
	glyph := "◇"
	switch notice.Level {
	case core.NoticeError:
		style = m.theme.Text.Error
		glyphStyle = m.theme.Text.Error
		glyph = "▲"
	case core.NoticeSuccess:
		style = m.theme.Text.Success
		glyphStyle = m.theme.Text.Success
		glyph = "•"
	}

	gutterWidth := noticeGutterWidth
	if gutterWidth < 1 {
		gutterWidth = 1
	}
	contentWidth := m.width - gutterWidth
	if contentWidth < 1 {
		contentWidth = 1
	}

	icon := lipgloss.NewStyle().
		Width(gutterWidth).
		AlignHorizontal(lipgloss.Left).
		Background(m.theme.Background).
		Render(glyphStyle.Render(glyph + " "))

	text := lipgloss.NewStyle().
		Width(contentWidth).
		Background(m.theme.Background).
		Render(style.Render(strings.TrimSpace(notice.Message)))

	return lipgloss.NewStyle().
		Background(m.theme.Background).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, icon, text))
}
