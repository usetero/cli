package root

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var (
	dialogConfirmBinding = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)
	dialogCancelBinding = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	)
	dialogToggleBinding = key.NewBinding(
		key.WithKeys("left", "right", "tab"),
		key.WithHelp("←/→", "select"),
	)
)

type quitDialogAction uint8

const (
	quitDialogNone quitDialogAction = iota
	quitDialogConfirm
	quitDialogCancel
)

type quitDialog struct {
	theme       theme.Theme
	selectedYes bool
}

func newQuitDialog(appTheme theme.Theme) *quitDialog {
	return &quitDialog{theme: appTheme}
}

func (d *quitDialog) Update(msg keyMsg) quitDialogAction {
	switch {
	case key.Matches(msg, immediateQuitBinding):
		return quitDialogConfirm
	case key.Matches(msg, dialogCancelBinding):
		return quitDialogCancel
	case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N"))):
		return quitDialogCancel
	case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y"))):
		return quitDialogConfirm
	case key.Matches(msg, dialogToggleBinding):
		d.selectedYes = !d.selectedYes
	case key.Matches(msg, dialogConfirmBinding):
		if d.selectedYes {
			return quitDialogConfirm
		}
		return quitDialogCancel
	}
	return quitDialogNone
}

func (d *quitDialog) ShortHelp() []key.Binding {
	return []key.Binding{
		dialogToggleBinding,
		dialogConfirmBinding,
		dialogCancelBinding,
		immediateQuitBinding,
	}
}

func (d *quitDialog) View() string {
	question := lipgloss.NewStyle().
		Foreground(d.theme.TextColor).
		Background(d.theme.Surface).
		Bold(true).
		Render("Are you sure you want to quit?")

	yes := d.renderButton("Yes", d.selectedYes)
	no := d.renderButton("No", !d.selectedYes)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		question,
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, yes, "  ", no),
	)

	return lipgloss.NewStyle().
		Background(d.theme.Surface).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.theme.Accent).
		BorderBackground(d.theme.Surface).
		Padding(1, 3).
		Render(content)
}

func (d *quitDialog) renderButton(label string, selected bool) string {
	if selected {
		return lipgloss.NewStyle().
			Background(d.theme.Accent).
			Foreground(d.theme.Surface).
			Padding(0, 2).
			Bold(true).
			Render(label)
	}
	return lipgloss.NewStyle().
		Background(d.theme.Surface).
		Foreground(d.theme.TextMuted).
		Padding(0, 2).
		Render(label)
}
