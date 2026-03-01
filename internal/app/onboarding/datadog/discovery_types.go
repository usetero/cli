package datadog

import (
	"time"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

const discoveryPollInterval = 500 * time.Millisecond

// pollTickMsg schedules the next async discovery status fetch.
type pollTickMsg struct{}

// statusMsg carries async discovery status fetch results.
type statusMsg struct {
	status *graphql.DatadogAccountStatus
	err    error
}
