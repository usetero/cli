package apitest

import (
	"context"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
)

// MockMessages is a mock implementation of api.Messages.
type MockMessages struct {
	CreateMessageFunc func(ctx context.Context, msg *domain.Message) error
}

var _ api.Messages = (*MockMessages)(nil)

// NewMockMessages creates a MockMessages with sensible defaults.
func NewMockMessages() *MockMessages {
	return &MockMessages{}
}

func (m *MockMessages) CreateMessage(ctx context.Context, msg *domain.Message) error {
	if m.CreateMessageFunc != nil {
		return m.CreateMessageFunc(ctx, msg)
	}
	return nil
}
