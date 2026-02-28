package datadog

import (
	"time"

	"github.com/usetero/cli/internal/api"
)

const discoveryPollInterval = 500 * time.Millisecond

// pollTickMsg schedules the next async discovery status fetch.
type pollTickMsg struct{}

// statusMsg carries async discovery status fetch results.
type statusMsg struct {
	status *api.DatadogAccountStatus
	err    error
}
