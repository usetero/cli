package datadogregion

import "github.com/usetero/cli/internal/domains/integrations"

// SelectedMsg reports a Datadog site choice from the region page.
type SelectedMsg struct {
	Site integrations.DatadogSite
}
