package msgs

import "github.com/usetero/cli/internal/domain"

// DatadogReady is emitted when datadog is already configured.
type DatadogReady struct{}

// DatadogNeeded is emitted when datadog setup is required.
type DatadogNeeded struct{}

// DatadogRegionSelected is emitted when user selects a datadog region.
type DatadogRegionSelected struct {
	Site domain.DatadogSite
}

// DatadogAPIKeyEntered is emitted when user enters API key.
type DatadogAPIKeyEntered struct {
	APIKey string
}

// DatadogAccountCreated is emitted when datadog account is created.
type DatadogAccountCreated struct {
	DatadogAccountID domain.DatadogAccountID
}

// DatadogDiscoveryComplete is emitted when discovery finishes.
type DatadogDiscoveryComplete struct{}
