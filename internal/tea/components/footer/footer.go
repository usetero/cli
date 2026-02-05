// Package footer provides a footer component with help text and error display.
package footer

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/keymap"
)

// Model displays help text and error state in the footer area.
type Model struct {
	theme  *styles.Theme
	help   help.Model
	keyMap help.KeyMap
	err    error
	width  int
}

// New creates a new footer.
func New(theme *styles.Theme) *Model {
	colors := theme.Colors

	h := help.New()
	h.Styles = help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		ShortDesc:      lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		ShortSeparator: lipgloss.NewStyle().Foreground(colors.BorderDefault),
		Ellipsis:       lipgloss.NewStyle().Foreground(colors.BorderDefault),
		FullKey:        lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		FullDesc:       lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		FullSeparator:  lipgloss.NewStyle().Foreground(colors.BorderDefault),
	}

	return &Model{
		theme: theme,
		help:  h,
	}
}

// SetWidth sets the footer width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.help.SetWidth(width)
}

// SetKeyBindings sets the key bindings to display.
func (m *Model) SetKeyBindings(bindings []key.Binding) {
	m.keyMap = keymap.Simple{Keys: bindings}
}

// SetError sets an error to display.
func (m *Model) SetError(err error) {
	m.err = err
}

// ClearError clears any displayed error.
func (m *Model) ClearError() {
	m.err = nil
}

// View renders the footer.
func (m *Model) View() string {
	var helpView string
	if m.keyMap != nil {
		helpView = m.help.ShortHelpView(m.keyMap.ShortHelp())
	}

	if m.err != nil {
		if helpView != "" {
			return lipgloss.JoinVertical(
				lipgloss.Left,
				m.renderError(),
				"",
				helpView,
			)
		}
		return m.renderError()
	}

	return helpView
}

// Height returns the rendered height of the footer.
func (m *Model) Height() int {
	return lipgloss.Height(m.View())
}

// renderError renders an error banner.
func (m *Model) renderError() string {
	colors := m.theme.Colors

	labelStyle := lipgloss.NewStyle().
		Background(colors.Error.Bg).
		Foreground(colors.Page.Text).
		Padding(0, 1).
		Bold(true)

	labelText := labelStyle.Render("ERROR")

	widthLeft := m.width - lipgloss.Width(labelText) - 2
	message := ansi.Truncate(m.err.Error(), widthLeft, "…")

	messageStyle := lipgloss.NewStyle().
		Background(colors.Error.Bg).
		Foreground(colors.Page.Text).
		Width(widthLeft+2).
		Padding(0, 1)

	messageText := messageStyle.Render(message)

	return ansi.Truncate(labelText+messageText, m.width, "…")
}
