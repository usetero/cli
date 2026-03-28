package core

// Screen is the shell-facing page contract.
type Screen interface {
	Model
	HelpProvider
	InputProvider
	BusyProvider
}
