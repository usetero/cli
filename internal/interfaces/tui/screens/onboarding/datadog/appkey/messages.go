package datadogappkey

import "github.com/usetero/cli/internal/domains/integrations"

const accountName = "Tero"

// SubmittedMsg reports a Datadog app-key submission.
type SubmittedMsg struct {
	AppKey string
}

func Submission(value string) integrations.DatadogAppKeySubmission {
	return integrations.DatadogAppKeySubmission{
		Name:   integrations.DatadogAccountName(accountName),
		AppKey: integrations.DatadogAppKey(value),
	}
}
