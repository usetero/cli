package errorfmt

import (
	"errors"
	"strings"
)

// UserFacing returns a concise, user-facing message derived from err.
// It preserves backend detail when available and falls back when needed.
func UserFacing(err error, fallback string) string {
	if err == nil {
		return fallback
	}

	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return fallback
	}

	// Unwrap and normalize common GraphQL wrapper noise.
	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		for _, e := range joined.Unwrap() {
			s := strings.TrimSpace(e.Error())
			if s != "" {
				msg = s
				break
			}
		}
	}

	msg = strings.TrimSpace(strings.TrimPrefix(msg, "graphql:"))
	if msg == "" {
		return fallback
	}

	return msg
}
