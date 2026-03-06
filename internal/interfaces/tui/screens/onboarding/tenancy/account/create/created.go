package accountcreate

import "github.com/usetero/cli/internal/domains/tenancy"

// CreatedMsg reports that the user submitted an account name to create.
type CreatedMsg struct {
	Create tenancy.AccountCreate
}
