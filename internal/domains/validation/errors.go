package validation

import "strings"

// FieldError is one validation failure.
type FieldError struct {
	Field string
	Rule  string
}

func (e FieldError) message() string {
	switch e.Rule {
	case "required", "notblank":
		return e.Field + " is required"
	case "max":
		return e.Field + " is too long"
	default:
		return e.Field + " is invalid"
	}
}

// Error contains one or more field validation failures.
type Error struct {
	Fields []FieldError
}

func (e *Error) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Fields))
	for _, fieldErr := range e.Fields {
		parts = append(parts, fieldErr.message())
	}
	return strings.Join(parts, "; ")
}
