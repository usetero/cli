package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	tuiapp "github.com/usetero/cli/internal/interfaces/tui/app"
	accountapp "github.com/usetero/cli/internal/interfaces/tui/app/account"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/filter"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type Dependencies struct {
	Scope                 logging.Scope
	Theme                 theme.Theme
	Environment           string
	Body                  core.Screen
	AccountRuntimeFactory accountapp.RuntimeFactory
}

func Start(deps Dependencies) error {
	if deps.Body == nil {
		panic("tui start requires body")
	}
	if deps.AccountRuntimeFactory == nil {
		panic("tui start requires account runtime factory")
	}

	program := tea.NewProgram(
		tuiapp.New(deps.Scope.Child("app"), deps.AccountRuntimeFactory, deps.Body, deps.Environment, deps.Theme),
		tea.WithEnvironment(os.Environ()),
		tea.WithFilter(filter.NewInputFilter()),
	)
	_, err := program.Run()
	return err
}
