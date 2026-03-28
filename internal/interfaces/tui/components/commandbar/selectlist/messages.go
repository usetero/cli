package selectlist

import "github.com/usetero/cli/internal/interfaces/tui/core"

// SelectedMsg is emitted when a command bar option is chosen.
type SelectedMsg struct {
	Option core.Option
}
