package account

import (
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// Scope identifies the organization/account context for a running account runtime.
type Scope struct {
	Organization tenancy.Organization
	Account      tenancy.Account
	Workspace    tenancy.Workspace
}

func (s Scope) Validate() error {
	if s.Organization.ID == "" {
		return fmt.Errorf("organization id is required")
	}
	if s.Account.ID == "" {
		return fmt.Errorf("account id is required")
	}
	return nil
}
