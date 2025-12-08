package datadogtest

import (
	"context"
)

// MockAPIKeyValidator is a mock implementation of datadog.DatadogAPIKeyValidator.
type MockAPIKeyValidator struct {
	ValidateAPIKeyFunc func(ctx context.Context, apiKey string, site string) (bool, string, error)
}

func (m *MockAPIKeyValidator) ValidateAPIKey(ctx context.Context, apiKey string, site string) (bool, string, error) {
	if m.ValidateAPIKeyFunc != nil {
		return m.ValidateAPIKeyFunc(ctx, apiKey, site)
	}
	return true, "", nil
}
