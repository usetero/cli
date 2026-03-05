package client

import "fmt"

// ErrorKind classifies client errors for retry policy.
type ErrorKind int

const (
	ErrorKindTransient ErrorKind = iota // Retry with backoff
	ErrorKindAuth                       // Refresh token and retry
	ErrorKindPermanent                  // Do not retry
)

// Error represents an error from the PowerSync service.
type Error struct {
	Kind       ErrorKind
	StatusCode int
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("powersync api: %d: %s", e.StatusCode, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("powersync api: %v", e.Err)
	}
	return fmt.Sprintf("powersync api: %s", e.Message)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) IsAuth() bool {
	return e != nil && e.Kind == ErrorKindAuth
}

func (e *Error) IsTransient() bool {
	return e != nil && e.Kind == ErrorKindTransient
}

func (e *Error) IsPermanent() bool {
	return e != nil && e.Kind == ErrorKindPermanent
}

func classifyHTTPStatus(statusCode int) ErrorKind {
	switch {
	case statusCode == 401 || statusCode == 403:
		return ErrorKindAuth
	case statusCode >= 500 || statusCode == 429:
		return ErrorKindTransient
	default:
		return ErrorKindPermanent
	}
}
