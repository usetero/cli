package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/palette"
)

// paletteCommands returns the list of commands available in the command palette.
func (m *Model) paletteCommands() []palette.Command {
	return []palette.Command{
		{
			Name: "New Conversation",
			Handler: func() tea.Cmd {
				// TODO: implement new conversation
				return nil
			},
		},
		{
			Name: "Toggle Details",
			Handler: func() tea.Cmd {
				m.statusBar.ToggleDrawer()
				return nil
			},
		},
		{
			Name: "Quit",
			Handler: func() tea.Cmd {
				m.quitDlg = newQuitDialog(m.theme)
				return nil
			},
		},
	}
}
