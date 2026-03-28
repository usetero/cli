package cli

import "github.com/usetero/cli/internal/interfaces/tui"

// TUI contains command-specific options for TUI mode.
type TUI struct{}

// Run starts TUI mode.
func (m *TUI) Run(exec *runner) error {
	deps, err := newTUIDependencies(exec)
	if err != nil {
		return err
	}
	return tui.Start(deps)
}
