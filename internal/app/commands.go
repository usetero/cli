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
				m.chat = m.newChat()
				m.statusBar.SetTitle("")
				m.windowTitle = ""
				m.updateLayout()
				return m.chat.Init()
			},
		},
		{
			Name: "Switch Organization",
			Handler: func() tea.Cmd {
				return m.switchOrganization()
			},
		},
		{
			Name: "Switch Account",
			Handler: func() tea.Cmd {
				return m.switchAccount()
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
