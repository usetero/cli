package apitest

import (
	"github.com/usetero/cli/internal/domain"
)

// NewOrganization creates a test organization with sensible defaults.
// Use functional options to override specific fields.
func NewOrganization(opts ...func(*domain.Organization)) domain.Organization {
	org := domain.Organization{
		ID:   domain.NewOrganizationID(),
		Name: "Test Organization",
	}
	for _, opt := range opts {
		opt(&org)
	}
	return org
}

// NewAccount creates a test account with sensible defaults.
// Use functional options to override specific fields.
func NewAccount(opts ...func(*domain.Account)) domain.Account {
	acc := domain.Account{
		ID:   domain.NewAccountID(),
		Name: "Test Account",
	}
	for _, opt := range opts {
		opt(&acc)
	}
	return acc
}
