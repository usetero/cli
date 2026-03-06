package datadogappkey

import "github.com/usetero/cli/internal/domains/integrations"

// SubmittedMsg reports that user submitted Datadog account name and app key.
type SubmittedMsg struct {
	Name   integrations.DatadogAccountName
	AppKey integrations.DatadogAppKey
}
