package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

// Start runs the TUI mode.
func Start(_ config.RuntimeConfig, scope logging.Scope) error {
	p := tea.NewProgram(New(scope), tea.WithEnvironment(os.Environ()))
	_, err := p.Run()
	return err
}
