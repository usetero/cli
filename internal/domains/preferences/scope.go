package preferences

import (
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/domains/validation"
)

// OrganizationSelection is the organization selection mutation input.
type OrganizationSelection struct {
	OrganizationID tenancy.OrganizationID `label:"organization id" validate:"required"`
}

// Validate validates organization selection input.
func (s OrganizationSelection) Validate() (OrganizationSelection, error) {
	if err := validation.Struct(s); err != nil {
		return OrganizationSelection{}, err
	}
	return s, nil
}

// AccountSelection is the account selection mutation input.
type AccountSelection struct {
	AccountID tenancy.AccountID `label:"account id" validate:"required"`
}

// Validate validates account selection input.
func (s AccountSelection) Validate() (AccountSelection, error) {
	if err := validation.Struct(s); err != nil {
		return AccountSelection{}, err
	}
	return s, nil
}

// WorkspaceSelection is the workspace selection mutation input.
type WorkspaceSelection struct {
	WorkspaceID tenancy.WorkspaceID `label:"workspace id" validate:"required"`
}

// Validate validates workspace selection input.
func (s WorkspaceSelection) Validate() (WorkspaceSelection, error) {
	if err := validation.Struct(s); err != nil {
		return WorkspaceSelection{}, err
	}
	return s, nil
}

// ScopeSelection is the full tenancy scope selection mutation input.
type ScopeSelection struct {
	OrganizationID tenancy.OrganizationID `label:"organization id" validate:"required"`
	AccountID      tenancy.AccountID      `label:"account id" validate:"required"`
	WorkspaceID    tenancy.WorkspaceID    `label:"workspace id" validate:"required"`
}

// Validate validates scope selection input.
func (s ScopeSelection) Validate() (ScopeSelection, error) {
	if err := validation.Struct(s); err != nil {
		return ScopeSelection{}, err
	}
	return s, nil
}

// Snapshot is the persisted preferences payload.
type Snapshot struct {
	Role         Role                   `json:"role,omitempty"`
	Organization tenancy.OrganizationID `json:"organization_id,omitempty"`
	Account      tenancy.AccountID      `json:"account_id,omitempty"`
	Workspace    tenancy.WorkspaceID    `json:"workspace_id,omitempty"`
}
