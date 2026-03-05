package db

import (
	"errors"
	"fmt"
)

// ErrCorrupt indicates a PowerSync local database invariant failure.
var ErrCorrupt = errors.New("powersync database corrupt")

// Error wraps store operations with context.
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("powersync db %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Op: op, Err: err}
}
