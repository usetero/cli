package cli

import "github.com/usetero/cli/internal/interfaces/tui"

// TUI contains command-specific options for TUI mode.
type TUI struct{}

// Run starts TUI mode.
func (m *TUI) Run(exec *runner) error {
	return tui.Start(exec.cfg, exec.scope.Child("tui"))
}
