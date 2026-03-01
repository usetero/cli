package datadog

import "github.com/usetero/cli/internal/domain"

// datadogAccountCreatedMsg is sent when Datadog account creation completes.
type datadogAccountCreatedMsg struct {
	datadogAccountID domain.DatadogAccountID
	err              error
}
