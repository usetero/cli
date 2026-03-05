package tenancy

import (
	"context"
)

type OrganizationID string

type Organization struct {
	ID                   OrganizationID
	Name                 string
	WorkosOrganizationID string
}

type OrganizationBootstrap struct {
	Organization Organization
	Account      Account
	Workspace    Workspace
}

// OrganizationService is the domain contract for organization operations.
type OrganizationService interface {
	List(ctx context.Context) ([]Organization, error)
	Create(ctx context.Context, name string) (OrganizationBootstrap, error)
}
