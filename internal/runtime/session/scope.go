package session

import (
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

// Scope identifies the organization/account context for a running session.
type Scope struct {
	OrganizationID tenancy.OrganizationID
	AccountID      tenancy.AccountID
}

func (s Scope) Validate() error {
	if s.OrganizationID == "" {
		return fmt.Errorf("organization id is required")
	}
	if s.AccountID == "" {
		return fmt.Errorf("account id is required")
	}
	return nil
}

// Status is the TUI-facing projection of current session lifecycle + sync state.
type Status struct {
	Running bool
	Scope   Scope
	Ready   bool
	Sync    pssyncer.State
}

type organizationScopedStorage interface {
	Storage
	SetOrganizationID(organizationID tenancy.OrganizationID)
}
