package powersynctest

import (
	"context"

	"github.com/usetero/cli/internal/powersync"
)

// Ensure MockTokenRefresher implements powersync.TokenRefresher.
var _ powersync.TokenRefresher = (*MockTokenRefresher)(nil)

// MockTokenRefresher is a test double for powersync.TokenRefresher.
type MockTokenRefresher struct {
	// GetAccessTokenFunc is called when GetAccessToken is invoked.
	GetAccessTokenFunc func(ctx context.Context) (string, error)

	// Calls records the number of times GetAccessToken was called.
	Calls int
}

// GetAccessToken implements powersync.TokenRefresher.
func (m *MockTokenRefresher) GetAccessToken(ctx context.Context) (string, error) {
	m.Calls++
	if m.GetAccessTokenFunc != nil {
		return m.GetAccessTokenFunc(ctx)
	}
	return "mock-token", nil
}

// NewMockTokenRefresher creates a MockTokenRefresher that returns the given token.
func NewMockTokenRefresher(token string) *MockTokenRefresher {
	return &MockTokenRefresher{
		GetAccessTokenFunc: func(ctx context.Context) (string, error) {
			return token, nil
		},
	}
}
