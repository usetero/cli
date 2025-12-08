package log_eventstest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockLogDiscoveryProgressPoller is a mock implementation of log_events.LogDiscoveryProgressPoller.
type MockLogDiscoveryProgressPoller struct {
	GetLogDiscoveryProgressFunc func(ctx context.Context, datadogAccountID string) (*api.LogEventDiscoveryProgress, error)
}

func (m *MockLogDiscoveryProgressPoller) GetLogDiscoveryProgress(ctx context.Context, datadogAccountID string) (*api.LogEventDiscoveryProgress, error) {
	if m.GetLogDiscoveryProgressFunc != nil {
		return m.GetLogDiscoveryProgressFunc(ctx, datadogAccountID)
	}
	return nil, nil
}
