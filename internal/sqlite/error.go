package sqlite

import (
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// SQLiteError wraps a sqlite3.Error with additional context.
// It provides structured access to error codes for debugging.
type SQLiteError struct {
	Code         int    // Primary error code
	ExtendedCode int    // Extended error code with more detail
	Message      string // Error message from sqlite3_errmsg()
	Op           string // Operation that failed (e.g., "insert message")
}

func (e *SQLiteError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %s (code=%d, extended=%d)", e.Op, e.Message, e.Code, e.ExtendedCode)
	}
	return fmt.Sprintf("%s (code=%d, extended=%d)", e.Message, e.Code, e.ExtendedCode)
}

// Unwrap returns nil since this is the root error.
func (e *SQLiteError) Unwrap() error {
	return nil
}

// WrapSQLiteError wraps an error with SQLite-specific details if available.
// If the error is not a sqlite3.Error, it returns a generic wrapped error.
func WrapSQLiteError(err error, op string) error {
	if err == nil {
		return nil
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return &SQLiteError{
			Code:         int(sqliteErr.Code),
			ExtendedCode: int(sqliteErr.ExtendedCode),
			Message:      err.Error(),
			Op:           op,
		}
	}

	// Not a SQLite error, wrap with context
	if op != "" {
		return fmt.Errorf("%s: %w", op, err)
	}
	return err
}
