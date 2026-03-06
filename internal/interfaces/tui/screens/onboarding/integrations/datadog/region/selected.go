package datadogregion

import "github.com/usetero/cli/internal/domains/integrations"

// SelectedMsg reports that the user confirmed a Datadog site.
type SelectedMsg struct {
	Site integrations.DatadogSite
}
