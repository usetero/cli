package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
)

// MockOrganizations implements api.Organizations for testing.
type MockOrganizations struct {
	ListFunc   func(ctx context.Context) ([]domain.Organization, error)
	CreateFunc func(ctx context.Context, input api.CreateOrganizationInput) (*api.OrganizationBootstrapResult, error)
}

// NewMockOrganizations creates a MockOrganizations with sensible defaults.
func NewMockOrganizations() *MockOrganizations {
	return &MockOrganizations{}
}

func (m *MockOrganizations) List(ctx context.Context) ([]domain.Organization, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockOrganizations) Create(ctx context.Context, input api.CreateOrganizationInput) (*api.OrganizationBootstrapResult, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}
