package datadogtest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockDatadogAccountCreator is a mock implementation of datadog.DatadogAccountCreator.
type MockDatadogAccountCreator struct {
	CreateAccountFunc func(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error)
}

func (m *MockDatadogAccountCreator) CreateAccount(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error) {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, accountID, name, site, apiKey, appKey)
	}
	return nil, nil
}
