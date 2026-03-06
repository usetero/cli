package organizationselect

import "github.com/usetero/cli/internal/domains/tenancy"

// SelectedMsg reports that the user confirmed an organization selection.
type SelectedMsg struct {
	OrganizationID tenancy.OrganizationID
}
