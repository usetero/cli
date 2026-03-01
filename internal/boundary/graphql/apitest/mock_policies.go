package apitest

import (
	"context"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

// MockPolicies is a mock implementation of graphql.Policies.
type MockPolicies struct {
	ApprovePolicyFunc func(ctx context.Context, policyID string) error
	DismissPolicyFunc func(ctx context.Context, policyID string) error
}

var _ graphql.Policies = (*MockPolicies)(nil)

// NewMockPolicies creates a MockPolicies with sensible defaults.
func NewMockPolicies() *MockPolicies {
	return &MockPolicies{}
}

func (m *MockPolicies) ApprovePolicy(ctx context.Context, policyID string) error {
	if m.ApprovePolicyFunc != nil {
		return m.ApprovePolicyFunc(ctx, policyID)
	}
	return nil
}

func (m *MockPolicies) DismissPolicy(ctx context.Context, policyID string) error {
	if m.DismissPolicyFunc != nil {
		return m.DismissPolicyFunc(ctx, policyID)
	}
	return nil
}
