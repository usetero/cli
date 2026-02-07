package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
)

// quitConfirmed is sent when the user confirms quitting.
type quitConfirmed struct{}

// quitCanceled is sent when the user dismisses the quit dialog.
type quitCanceled struct{}

// quitDialog is a quit confirmation dialog with Yes/No buttons.
type quitDialog struct {
	theme       *styles.Theme
	selectedYes bool
}

// newQuitDialog creates a new quit dialog. Defaults to "No" selected.
func newQuitDialog(theme *styles.Theme) *quitDialog {
	return &quitDialog{theme: theme}
}

// Update handles key presses and returns the user's choice as a message.
func (m *quitDialog) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return func() tea.Msg { return quitConfirmed{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y"))):
			return func() tea.Msg { return quitConfirmed{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N", "esc"))):
			return func() tea.Msg { return quitCanceled{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "right", "tab"))):
			m.selectedYes = !m.selectedYes
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.selectedYes {
				return func() tea.Msg { return quitConfirmed{} }
			}
			return func() tea.Msg { return quitCanceled{} }
		}
	}
	return nil
}

// View renders the dialog box. The caller positions it via compositor layers.
func (m *quitDialog) View() string {
	colors := m.theme.Colors

	question := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Bold(true).
		Render("Are you sure you want to quit?")

	yesBtn := m.renderButton("Yes", m.selectedYes)
	noBtn := m.renderButton("No", !m.selectedYes)
	buttons := yesBtn + "  " + noBtn

	content := lipgloss.JoinVertical(lipgloss.Center, question, "", buttons)

	return lipgloss.NewStyle().
		Background(colors.Page.Bg).
		Foreground(colors.Page.Text).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Accent).
		BorderBackground(colors.Page.Bg).
		Padding(1, 3).
		Render(content)
}

// renderButton renders a single button with selected/unselected styling.
func (m *quitDialog) renderButton(label string, selected bool) string {
	colors := m.theme.Colors

	if selected {
		return lipgloss.NewStyle().
			Background(colors.Accent).
			Foreground(colors.Page.Bg).
			Padding(0, 3).
			Bold(true).
			Render(label)
	}

	return lipgloss.NewStyle().
		Background(colors.Page.Bg).
		Foreground(colors.Page.TextMuted).
		Padding(0, 3).
		Render(label)
}
