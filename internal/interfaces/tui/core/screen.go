package core

// Screen is the shell-facing page contract.
type Screen interface {
	Model
	PageProvider
	CommandProvider
	HelpProvider
	InputProvider
	BusyProvider
	ErrorProvider
	NoticeProvider
}
