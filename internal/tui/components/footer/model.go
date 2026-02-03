package footer

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
)

// Model displays help text and error state.
type Model struct {
	theme  *styles.Theme
	logger log.Logger
	help   help.Model
	keyMap help.KeyMap
	err    error
	width  int
}

// New creates a new footer model.
func New(theme *styles.Theme, logger log.Logger) Model {
	colors := theme.Colors
	helpModel := help.New()
	helpModel.Styles = help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		ShortDesc:      lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		ShortSeparator: lipgloss.NewStyle().Foreground(colors.BorderDefault),
		Ellipsis:       lipgloss.NewStyle().Foreground(colors.BorderDefault),
		FullKey:        lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		FullDesc:       lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		FullSeparator:  lipgloss.NewStyle().Foreground(colors.BorderDefault),
	}

	return Model{
		theme:  theme,
		logger: logger,
		help:   helpModel,
	}
}

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	m.help.SetWidth(width)
	return m
}

// SetKeyBindings returns a new Model with the given key bindings.
func (m Model) SetKeyBindings(bindings []key.Binding) Model {
	m.keyMap = keymap.Simple{Keys: bindings}
	return m
}

// SetError returns a new Model with the given error.
func (m Model) SetError(err error) Model {
	m.err = err
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the footer.
func (m Model) View() string {
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

// renderError renders an error banner.
func (m Model) renderError() string {
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

// Height returns the rendered height of the footer.
func (m Model) Height() int {
	return lipgloss.Height(m.View())
}
