package tenancy

import (
	"context"
	"strings"

	"github.com/usetero/cli/internal/domains/validation"
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

// OrganizationCreate is the organization creation mutation input.
type OrganizationCreate struct {
	Name string `label:"organization name" validate:"required,notblank,max=100"`
}

// Validate normalizes and validates organization create input.
func (c OrganizationCreate) Validate() (OrganizationCreate, error) {
	c.Name = strings.TrimSpace(c.Name)
	if err := validation.Struct(c); err != nil {
		return OrganizationCreate{}, err
	}
	return c, nil
}

// OrganizationService is the domain contract for organization operations.
type OrganizationService interface {
	List(ctx context.Context) ([]Organization, error)
	Create(ctx context.Context, create OrganizationCreate) (OrganizationBootstrap, error)
}
