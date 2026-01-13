package datadogtest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockStatusPoller implements datadog.StatusPoller for testing.
type MockStatusPoller struct {
	GetStatusFunc func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error)
}

func (m *MockStatusPoller) GetStatus(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, datadogAccountID)
	}
	return nil, nil
}
