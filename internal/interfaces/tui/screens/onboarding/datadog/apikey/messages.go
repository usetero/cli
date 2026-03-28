package datadogapikey

import "github.com/usetero/cli/internal/domains/integrations"

// SubmittedMsg reports a Datadog API-key submission.
type SubmittedMsg struct {
	APIKey string
}

func Submission(value string) integrations.DatadogAPIKeySubmission {
	return integrations.DatadogAPIKeySubmission{
		APIKey: integrations.DatadogAPIKey(value),
	}
}
