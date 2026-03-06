package providerselect

import "github.com/usetero/cli/internal/domains/integrations"

// SelectedMsg reports that the user selected an integration provider.
type SelectedMsg struct {
	Provider integrations.Provider
}
