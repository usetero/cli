package integrationsflow

import "github.com/usetero/cli/internal/domains/integrations"

type ProviderSelectedMsg struct {
	Provider integrations.Provider
}

type SetDatadogSiteMsg struct {
	Site integrations.DatadogSite
}

type SubmitDatadogAPIKeyMsg struct {
	Submission integrations.DatadogAPIKeySubmission
}

type SubmitDatadogAppKeyMsg struct {
	Submission integrations.DatadogAppKeySubmission
}

type RefreshRequestedMsg struct{}
