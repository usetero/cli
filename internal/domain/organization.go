package domain

import "github.com/google/uuid"

// OrganizationID is a unique identifier for an organization.
type OrganizationID string

// NewOrganizationID generates a new unique OrganizationID.
func NewOrganizationID() OrganizationID { return OrganizationID(uuid.New().String()) }

func (id OrganizationID) String() string { return string(id) }

// WorkosOrganizationID is the external ID from WorkOS.
type WorkosOrganizationID string

func (id WorkosOrganizationID) String() string { return string(id) }

// Organization represents a customer organization.
type Organization struct {
	ID                   OrganizationID       `json:"id"`
	Name                 string               `json:"name"`
	WorkosOrganizationID WorkosOrganizationID `json:"workos_organization_id,omitempty"`
}
