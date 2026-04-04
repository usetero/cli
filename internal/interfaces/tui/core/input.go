package core

// InputKind is the kind of page-owned input currently requested.
type InputKind uint8

const (
	InputNone InputKind = iota
	InputConfirm
	InputText
	InputMultiline
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
	Title       string
	Detail      string
	Label       string
	Action      string
	Placeholder string
	Secret      bool
	Options     []Option
}

// InputProvider exposes the current generic shell input requested by a model.
type InputProvider interface {
	Input() *Input
}
