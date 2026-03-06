package datadogapikey

import "github.com/usetero/cli/internal/domains/integrations"

// SubmittedMsg reports that the user submitted a Datadog API key.
type SubmittedMsg struct {
	APIKey integrations.DatadogAPIKey
}
