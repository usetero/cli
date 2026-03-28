package core

import "charm.land/bubbles/v2/key"

// InputKind is the kind of page-owned input currently requested.
type InputKind uint8

const (
	InputNone InputKind = iota
	InputAction
	InputText
	InputSelect
)

// Option is one generic selectable input option.
type Option struct {
	ID       string
	Label    string
	Subtitle string
}

// Input describes the current generic input a page wants from the shell.
type Input struct {
	Kind        InputKind
	Label       string
	Action      string
	Placeholder string
	Options     []Option
}

// HelpProvider exposes key bindings handled by a model.
type HelpProvider interface {
	ShortHelp() []key.Binding
}

// InputProvider exposes the current generic shell input requested by a model.
type InputProvider interface {
	Input() *Input
}
