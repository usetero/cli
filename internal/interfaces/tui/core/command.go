package core

// CommandID identifies a shell command.
type CommandID string

const (
	CommandQuit         CommandID = "quit"
	CommandOpenServices CommandID = "open_services"
	CommandOpenSpikes   CommandID = "open_spikes"
	CommandOpenWaste    CommandID = "open_waste"
)

// Command is one shell command exposed to the command palette.
type Command struct {
	ID          CommandID
	Title       string
	Description string
}

// CommandProvider exposes shell commands available in the current context.
type CommandProvider interface {
	Commands() []Command
}

// CommandSelectedMsg is emitted when the user selects a shell command.
type CommandSelectedMsg struct {
	ID CommandID
}
