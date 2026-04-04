package palette

import "github.com/usetero/cli/internal/interfaces/tui/core"

// SubmittedMsg is emitted when the user chooses a palette command.
type SubmittedMsg struct {
	Command core.Command
}
