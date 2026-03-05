package preferences

import (
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// Role is the user-selected onboarding role.
type Role string

const (
	RolePlatform Role = "platform"
	RoleEngineer Role = "engineer"
)

func (r Role) Valid() bool {
	switch r {
	case RolePlatform, RoleEngineer:
		return true
	default:
		return false
	}
}

func (r Role) String() string {
	return string(r)
}

func ParseRole(raw string) (Role, error) {
	r := Role(raw)
	if !r.Valid() {
		return "", fmt.Errorf("invalid role: %q", raw)
	}
	return r, nil
}

// Snapshot is the persisted preferences payload.
type Snapshot struct {
	Role         Role                   `json:"role,omitempty"`
	Organization tenancy.OrganizationID `json:"organization_id,omitempty"`
	Account      tenancy.AccountID      `json:"account_id,omitempty"`
	Workspace    tenancy.WorkspaceID    `json:"workspace_id,omitempty"`
}
