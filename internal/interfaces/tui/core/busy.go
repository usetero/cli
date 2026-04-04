package core

// Progress describes optional progress metadata for a busy state.
type Progress struct {
	Current int
	Total   int
}

// Busy describes a busy shell state that overrides active input.
type Busy struct {
	Label    string
	Detail   string
	Status   string
	Progress *Progress
}

// BusyProvider exposes the current busy state for a model.
type BusyProvider interface {
	Busy() *Busy
}
