package preferences

import (
	"fmt"

	"github.com/usetero/cli/internal/domains/validation"
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

// RoleSelection is the role selection mutation input.
type RoleSelection struct {
	Role Role
}

// Validate validates role selection input.
func (s RoleSelection) Validate() (RoleSelection, error) {
	if err := validation.Struct(struct {
		Role Role `label:"role" validate:"required"`
	}{Role: s.Role}); err != nil {
		return RoleSelection{}, err
	}
	if !s.Role.Valid() {
		return RoleSelection{}, fmt.Errorf("invalid role: %q", s.Role)
	}
	return s, nil
}
