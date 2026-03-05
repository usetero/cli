package syncer

import "errors"

var (
	// ErrAlreadyStarted indicates Start was called while syncer is already running.
	ErrAlreadyStarted = errors.New("powersync syncer already started")
	// ErrInvalidInput indicates a required input was missing.
	ErrInvalidInput = errors.New("powersync syncer invalid input")
)
