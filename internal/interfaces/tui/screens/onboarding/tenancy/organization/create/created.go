package organizationcreate

import "github.com/usetero/cli/internal/domains/tenancy"

// CreatedMsg reports that the user submitted an organization name to create.
type CreatedMsg struct {
	Create tenancy.OrganizationCreate
}
