package auth

import (
	"time"

	"github.com/usetero/cli/internal/auth"
)

type authState int

const (
	stateInitializing authState = iota
	stateReady
	statePolling
	stateComplete
)

const (
	defaultPollInterval = 2 * time.Second
	minPollInterval     = 1 * time.Second
	maxPollInterval     = 2 * time.Second
)

// deviceAuthStartedMsg is sent when device auth flow starts.
type deviceAuthStartedMsg struct {
	deviceAuth *auth.DeviceAuth
	err        error
}

// authCompletedMsg is sent when auth completes.
type authCompletedMsg struct {
	result *auth.Result
	err    error
}
