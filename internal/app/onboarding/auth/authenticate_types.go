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

// deviceAuthMsg is sent when device auth flow starts.
type deviceAuthMsg struct {
	deviceAuth *auth.DeviceAuth
	err        error
}

// authCompleteMsg is sent when auth completes.
type authCompleteMsg struct {
	result *auth.Result
	err    error
}
