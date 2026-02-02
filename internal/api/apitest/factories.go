package apitest

import (
	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
)

// NewOrganization creates a test organization with sensible defaults.
// Use functional options to override specific fields.
func NewOrganization(opts ...func(*api.Organization)) api.Organization {
	org := api.Organization{
		ID:   uuid.New().String(),
		Name: "Test Organization",
	}
	for _, opt := range opts {
		opt(&org)
	}
	return org
}

// NewAccount creates a test account with sensible defaults.
// Use functional options to override specific fields.
func NewAccount(opts ...func(*api.Account)) api.Account {
	acc := api.Account{
		ID:   uuid.New().String(),
		Name: "Test Account",
	}
	for _, opt := range opts {
		opt(&acc)
	}
	return acc
}

// NewWorkspace creates a test workspace with sensible defaults.
// Use functional options to override specific fields.
func NewWorkspace(opts ...func(*api.Workspace)) api.Workspace {
	ws := api.Workspace{
		ID:   uuid.New().String(),
		Name: "Test Workspace",
	}
	for _, opt := range opts {
		opt(&ws)
	}
	return ws
}
