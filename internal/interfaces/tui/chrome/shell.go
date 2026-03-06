package chrome

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Viewport is the current terminal viewport size.
type Viewport struct {
	Width  int
	Height int
}

// Slots defines the shell content regions.
type Slots struct {
	Header string
	Body   string
	Footer string
}

// Render draws the shared app shell around body content.
func Render(t theme.Theme, slots Slots, viewport Viewport) tea.View {
	body := strings.TrimRight(slots.Body, "\n")
	if body == "" {
		body = " "
	}
	header := strings.TrimSpace(slots.Header)
	if header == "" {
		header = t.Shell.HeaderLead.Render(" ")
	}
	footer := strings.TrimSpace(slots.Footer)

	headerLine := t.Shell.HeaderBar.Render(header)
	footerLine := t.Shell.Footer.Render(footer)
	bodyBlock := t.Shell.Body.Render(body)
	sections := []string{headerLine}

	if viewport.Width > 0 && viewport.Height > 0 {
		innerWidth := viewport.Width - t.Shell.Outer.GetHorizontalFrameSize()
		if innerWidth < 0 {
			innerWidth = 0
		}
		innerHeight := viewport.Height - t.Shell.Outer.GetVerticalFrameSize()
		if innerHeight < 0 {
			innerHeight = 0
		}

		headerLine = t.Shell.HeaderBar.Width(innerWidth).Render(header)
		footerLine = t.Shell.Footer.Width(innerWidth).Render(footer)
		bodyContent := t.Shell.Body.Width(innerWidth).Render(body)

		headerHeight := lipgloss.Height(headerLine)
		footerHeight := 0
		if footer != "" {
			footerHeight = lipgloss.Height(footerLine)
		}
		bodyHeight := innerHeight - headerHeight - footerHeight
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		bodyBlock = lipgloss.NewStyle().Width(innerWidth).Height(bodyHeight).Render(bodyContent)
	}

	sections = append(sections, bodyBlock)
	if footer != "" {
		sections = append(sections, footerLine)
	}

	rendered := t.Shell.Outer.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
	if viewport.Width > 0 && viewport.Height > 0 {
		rendered = t.Shell.Outer.Width(viewport.Width).Height(viewport.Height).Render(
			lipgloss.JoinVertical(lipgloss.Left, sections...),
		)
	}
	return tea.NewView(rendered)
}
