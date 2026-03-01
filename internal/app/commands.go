package app

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/app/palette"
	"github.com/usetero/cli/internal/preferences"
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
			Name:     "Theme",
			Children: m.themeCommands(),
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

// themeCommands returns sub-commands for theme selection.
func (m *Model) themeCommands() []palette.Command {
	set := func(theme preferences.Theme) func() tea.Cmd {
		return func() tea.Cmd {
			return func() tea.Msg { return appevents.ThemeChangeRequested{Theme: theme} }
		}
	}
	return []palette.Command{
		{Name: "Auto (detect from terminal)", Handler: set(preferences.ThemeAuto)},
		{Name: "Dark", Handler: set(preferences.ThemeDark)},
		{Name: "Light", Handler: set(preferences.ThemeLight)},
	}
}
