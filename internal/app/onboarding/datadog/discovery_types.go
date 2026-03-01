package datadog

import (
	"time"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

const discoveryPollInterval = 500 * time.Millisecond

// discoveryPollTickMsg schedules the next async discovery status fetch.
type discoveryPollTickMsg struct{}

// discoveryStatusMsg carries async discovery status fetch results.
type discoveryStatusMsg struct {
	status *graphql.DatadogAccountStatus
	err    error
}
