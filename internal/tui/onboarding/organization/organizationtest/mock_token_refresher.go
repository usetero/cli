package organizationtest

import (
	"context"
)

// MockTokenRefresher implements organization.TokenRefresher for testing.
type MockTokenRefresher struct {
	RefreshTokenWithOrganizationFunc func(ctx context.Context, workosOrgID string) (string, error)
}

func (m *MockTokenRefresher) RefreshTokenWithOrganization(ctx context.Context, workosOrgID string) (string, error) {
	if m.RefreshTokenWithOrganizationFunc != nil {
		return m.RefreshTokenWithOrganizationFunc(ctx, workosOrgID)
	}
	return "", nil
}
