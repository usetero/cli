package app

import (
	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

type children struct {
	core.Children

	statusbar  *statusbar.Model
	commandbar *commandbar.Model
	helpbar    *helpbar.Model
}

func newChildren(status *statusbar.Model, command *commandbar.Model, help *helpbar.Model) children {
	coreChildren := core.Children{status, command, help}
	return children{
		Children:   coreChildren,
		statusbar:  status,
		commandbar: command,
		helpbar:    help,
	}
}

func (c children) refreshFooter(body core.Screen) {
	c.commandbar.SetState(body.Input(), body.Busy())

	var bindings []key.Binding
	bindings = append(bindings, body.ShortHelp()...)
	bindings = append(bindings, c.commandbar.ShortHelp()...)
	bindings = append(bindings, quitBinding)
	c.helpbar.SetBindings(bindings)
}
