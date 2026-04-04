package app

import (
	"charm.land/bubbles/v2/key"
	chromedivider "github.com/usetero/cli/internal/interfaces/tui/components/chrome/divider"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type shell struct {
	core.Children

	statusbar  *statusbar.Model
	divider    *chromedivider.Model
	commandbar *commandbar.Model
	helpbar    *helpbar.Model
}

func newShell(status *statusbar.Model, divider *chromedivider.Model, command *commandbar.Model, help *helpbar.Model) shell {
	coreChildren := core.Children{status, divider, command, help}
	return shell{
		Children:   coreChildren,
		statusbar:  status,
		divider:    divider,
		commandbar: command,
		helpbar:    help,
	}
}

func (s shell) UpdateState(body core.Screen) {
	s.commandbar.SetState(body.Input(), body.Busy(), body.Error(), body.Notice())
	commands := append([]core.Command{
		{
			ID:          core.CommandQuit,
			Title:       "Quit",
			Description: "Close Tero",
		},
	}, body.Commands()...)
	s.commandbar.SetCommands(commands)
	title := theme.AppName
	if page := body.Page(); page.Title != "" {
		title = s.commandbar.FooterTitle(page.Title)
	}
	s.divider.SetState(title, body.Busy() != nil, s.commandbar.IsPaletteOpen(), 0)

	var bindings []key.Binding
	bindings = append(bindings, body.ShortHelp()...)
	bindings = append(bindings, s.commandbar.ShortHelp()...)
	bindings = append(bindings, quitBinding)
	s.helpbar.SetBindings(bindings)
}
