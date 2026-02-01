package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockOrganizations implements api.Organizations for testing.
type MockOrganizations struct {
	ListFunc   func(ctx context.Context) ([]api.Organization, error)
	CreateFunc func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error)
}

func (m *MockOrganizations) List(ctx context.Context) ([]api.Organization, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockOrganizations) Create(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, name)
	}
	return nil, nil
}
