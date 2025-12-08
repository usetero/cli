package organizationtest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockOrgCreator implements organization.OrganizationCreator for testing.
type MockOrgCreator struct {
	CreateFunc func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error)
}

func (m *MockOrgCreator) Create(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, name)
	}
	return nil, nil
}
