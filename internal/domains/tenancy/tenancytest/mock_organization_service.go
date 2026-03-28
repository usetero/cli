package tenancytest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// MockOrganizationService is a functional mock for tenancy.OrganizationService.
type MockOrganizationService struct {
	ListFn   func(ctx context.Context) ([]tenancy.Organization, error)
	CreateFn func(ctx context.Context, create tenancy.OrganizationCreate) (tenancy.OrganizationBootstrap, error)
}

var _ tenancy.OrganizationService = (*MockOrganizationService)(nil)

func NewMockOrganizationService() *MockOrganizationService {
	return &MockOrganizationService{}
}

func (m *MockOrganizationService) List(ctx context.Context) ([]tenancy.Organization, error) {
	if m.ListFn == nil {
		return nil, nil
	}
	return m.ListFn(ctx)
}

func (m *MockOrganizationService) Create(ctx context.Context, create tenancy.OrganizationCreate) (tenancy.OrganizationBootstrap, error) {
	if m.CreateFn == nil {
		return tenancy.OrganizationBootstrap{}, nil
	}
	return m.CreateFn(ctx, create)
}
