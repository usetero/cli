package cli

import "github.com/usetero/cli/internal/interfaces/tui"

// Run starts the default CLI mode (TUI).
func (c *CLI) Run(exec *runner) error {
	return tui.Start(exec.cfg, exec.scope.Child("tui"))
}
