package graphql

import (
	"errors"
	"strings"
)

// Sentinel errors for API responses.
// Use errors.Is() to check for these error types.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates the resource already exists.
	ErrAlreadyExists = errors.New("already exists")
)

// notFoundPatterns are substrings that indicate a "not found" error from the API.
var notFoundPatterns = []string{
	"not found",
	"not_found",
	"does not exist",
}

// alreadyExistsPatterns are substrings that indicate an "already exists" error from the API.
var alreadyExistsPatterns = []string{
	"already exists",
	"already_exists",
	"duplicate",
	"conflict",
}

// classifyError examines an error message and returns a wrapped sentinel error
// if the message matches known patterns. Returns nil if no pattern matches.
func classifyError(err error) error {
	if err == nil {
		return nil
	}

	msg := strings.ToLower(err.Error())

	for _, pattern := range notFoundPatterns {
		if strings.Contains(msg, pattern) {
			return ErrNotFound
		}
	}

	for _, pattern := range alreadyExistsPatterns {
		if strings.Contains(msg, pattern) {
			return ErrAlreadyExists
		}
	}

	return nil
}
