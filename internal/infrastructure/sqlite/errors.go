package sqlite

import "fmt"

// Error wraps low-level SQLite operations with context.
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("sqlite %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Op: op, Err: err}
}
