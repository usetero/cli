package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
)

// MockAPIServiceServices is a mock implementation of api.Services.
type MockAPIServiceServices struct {
	EnableServiceFunc  func(ctx context.Context, serviceID string) error
	DisableServiceFunc func(ctx context.Context, serviceID string) error
}

var _ api.Services = (*MockAPIServiceServices)(nil)

// NewMockAPIServiceServices creates a MockAPIServiceServices with sensible defaults.
func NewMockAPIServiceServices() *MockAPIServiceServices {
	return &MockAPIServiceServices{}
}

func (m *MockAPIServiceServices) EnableService(ctx context.Context, serviceID string) error {
	if m.EnableServiceFunc != nil {
		return m.EnableServiceFunc(ctx, serviceID)
	}
	return nil
}

func (m *MockAPIServiceServices) DisableService(ctx context.Context, serviceID string) error {
	if m.DisableServiceFunc != nil {
		return m.DisableServiceFunc(ctx, serviceID)
	}
	return nil
}
