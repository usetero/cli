package apitest

import (
	"context"
)

// MockPolicies implements api.LogEventPolicies for testing.
type MockPolicies struct {
	ApproveFunc func(ctx context.Context, id string) error
	DismissFunc func(ctx context.Context, id string) error
}

// NewMockPolicies creates a MockPolicies with sensible defaults.
func NewMockPolicies() *MockPolicies {
	return &MockPolicies{}
}

func (m *MockPolicies) Approve(ctx context.Context, id string) error {
	if m.ApproveFunc != nil {
		return m.ApproveFunc(ctx, id)
	}
	return nil
}

func (m *MockPolicies) Dismiss(ctx context.Context, id string) error {
	if m.DismissFunc != nil {
		return m.DismissFunc(ctx, id)
	}
	return nil
}
