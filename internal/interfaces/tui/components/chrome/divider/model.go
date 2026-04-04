package divider

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

var leadFrames = []string{"◇", "◈", "◆", "◈"}

const footerTitleSuffix = " [/]"
const dividerPadRight = 2

// Model renders shell divider chrome above the command surface.
type Model struct {
	theme       theme.Theme
	width       int
	enabled     bool
	title       string
	busy        bool
	paletteOpen bool
	spinnerStep int
}

var _ core.Model = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{
		theme:   appTheme,
		enabled: true,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

func (m *Model) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func (m *Model) SetState(title string, busy bool, paletteOpen bool, spinnerStep int) {
	m.title = title
	m.busy = busy
	m.paletteOpen = paletteOpen
	m.spinnerStep = spinnerStep
}

func (m *Model) View() tea.View {
	if !m.enabled || m.width <= 0 {
		return tea.NewView("")
	}

	lead := "◆"
	leadStyle := lipgloss.NewStyle().
		Foreground(m.theme.Palette.Brand).
		Background(m.theme.Background).
		Bold(true)
	if m.busy {
		lead = leadFrames[m.spinnerStep%len(leadFrames)]
		leadColor := m.theme.Palette.Brand
		if m.spinnerStep%2 == 1 {
			leadColor = m.theme.Palette.Accent
		}
		leadStyle = lipgloss.NewStyle().
			Foreground(leadColor).
			Background(m.theme.Background).
			Bold(true)
	}
	if m.paletteOpen {
		lead = "‹"
		leadStyle = m.theme.Shell.FooterLead
	}

	title := m.title
	if title == "" {
		title = theme.AppName
	}

	label := leadStyle.Render(lead) + " " + m.renderTitle(title)
	lineWidth := m.width - dividerPadRight - lipgloss.Width(label) - 1
	if lineWidth < 0 {
		lineWidth = 0
	}
	rule := m.theme.Shell.FooterRule.Render(lipgloss.NewStyle().
		Background(m.theme.Background).
		Render(strings.Repeat("─", lineWidth)))

	return tea.NewView(lipgloss.NewStyle().
		Width(m.width).
		PaddingRight(dividerPadRight).
		Background(m.theme.Background).
		Render(label + " " + rule))
}

func (m *Model) renderTitle(title string) string {
	if m.paletteOpen {
		return m.theme.Shell.FooterLead.Render(title)
	}

	if strings.HasSuffix(title, footerTitleSuffix) {
		pageTitle := strings.TrimSuffix(title, footerTitleSuffix)
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().
				Foreground(m.theme.Palette.Brand).
				Background(m.theme.Background).
				Bold(true).
				Render(pageTitle),
			m.theme.Shell.FooterLead.Render(footerTitleSuffix),
		)
	}

	return lipgloss.NewStyle().
		Foreground(m.theme.Palette.Brand).
		Background(m.theme.Background).
		Bold(true).
		Render(title)
}
