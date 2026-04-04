package core

// Error describes a blocking shell error state.
type Error struct {
	Message string
	Detail string
	Action string
}

// ErrorProvider exposes the current error state for a model.
type ErrorProvider interface {
	Error() *Error
}
