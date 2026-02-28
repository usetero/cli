package datadog

import "github.com/usetero/cli/internal/domain"

// accountCreatedMsg is sent when account creation completes.
type accountCreatedMsg struct {
	datadogAccountID domain.DatadogAccountID
	err              error
}
