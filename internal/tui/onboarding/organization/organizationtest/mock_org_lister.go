package organizationtest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockOrgLister implements organization.OrganizationLister for testing.
type MockOrgLister struct {
	ListFunc func(ctx context.Context) ([]api.Organization, error)
}

func (m *MockOrgLister) List(ctx context.Context) ([]api.Organization, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}
