package organizationselect

import "github.com/usetero/cli/internal/domains/tenancy"

// SelectedMsg reports an organization choice from the select page.
type SelectedMsg struct {
	OrganizationID tenancy.OrganizationID
}
