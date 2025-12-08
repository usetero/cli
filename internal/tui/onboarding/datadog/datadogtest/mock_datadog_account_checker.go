package datadogtest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockDatadogAccountChecker is a mock implementation of datadog.DatadogAccountChecker.
type MockDatadogAccountChecker struct {
	HasAccountFunc func(ctx context.Context, accountID string) (bool, error)
	GetAccountFunc func(ctx context.Context, accountID string) (*api.DatadogAccount, error)
}

func (m *MockDatadogAccountChecker) HasAccount(ctx context.Context, accountID string) (bool, error) {
	if m.HasAccountFunc != nil {
		return m.HasAccountFunc(ctx, accountID)
	}
	return false, nil
}

func (m *MockDatadogAccountChecker) GetAccount(ctx context.Context, accountID string) (*api.DatadogAccount, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, accountID)
	}
	return nil, nil
}
