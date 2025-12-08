package servicestest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockServiceDiscoveryPoller is a mock implementation of services.ServiceDiscoveryPoller.
type MockServiceDiscoveryPoller struct {
	GetServiceDiscoveryStatusFunc func(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error)
}

func (m *MockServiceDiscoveryPoller) GetServiceDiscoveryStatus(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error) {
	if m.GetServiceDiscoveryStatusFunc != nil {
		return m.GetServiceDiscoveryStatusFunc(ctx, datadogAccountID)
	}
	return nil, nil
}
