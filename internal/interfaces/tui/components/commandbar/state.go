package commandbar

// Mode is the active interaction state for the global command bar.
type Mode int

const (
	ModeHidden Mode = iota
	ModeAction
	ModeInput
	ModeSelect
)
