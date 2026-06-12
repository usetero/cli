package graphql

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/log"
)

type Policies interface {
	// ApprovePolicy approves a policy with the given ID.
	ApprovePolicy(ctx context.Context, policyID string) error
	// DismissPolicy dismisses a policy with the given ID.
	DismissPolicy(ctx context.Context, policyID string) error
}

type PolicyService struct {
	client Client
	scope  log.Scope
}

// Ensure PolicyService implements Policies.
var _ Policies = (*PolicyService)(nil)

// NewPolicyService creates a new policy service.
func NewPolicyService(client Client, scope log.Scope) *PolicyService {
	return &PolicyService{
		client: client,
		scope:  scope.Child("policies"),
	}
}

// ApprovePolicy approves a log event policy.
//
// TODO(drop-powersync): the policy lifecycle moved to the Issue model in the
// control plane (createLogEventPolicy / ignoreIssue). Re-wire as an inline
// mutation in the writes step (task #5).
func (s *PolicyService) ApprovePolicy(_ context.Context, id string) error {
	return fmt.Errorf("approve policy %s: not wired — moved to the issue model", id)
}

// DismissPolicy dismisses a log event policy.
//
// TODO(drop-powersync): see ApprovePolicy — moved to the Issue model.
func (s *PolicyService) DismissPolicy(_ context.Context, id string) error {
	return fmt.Errorf("dismiss policy %s: not wired — moved to the issue model", id)
}
