package datadog

import (
	"time"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

const discoveryPollInterval = 500 * time.Millisecond

// discoveryPollTickMsg schedules the next async discovery status fetch.
type discoveryPollTickMsg struct{}

// discoveryStatusLoadedMsg carries async discovery status fetch results.
type discoveryStatusLoadedMsg struct {
	status *graphql.DatadogAccountStatus
	err    error
}
