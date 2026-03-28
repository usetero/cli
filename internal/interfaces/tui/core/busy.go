package core

// Busy describes a busy shell state that overrides active input.
type Busy struct {
	Label  string
	Detail string
}

// BusyProvider exposes the current busy state for a model.
type BusyProvider interface {
	Busy() *Busy
}
