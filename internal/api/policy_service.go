package api

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/log"
)

// LogEventPolicies provides access to log event policy operations.
type LogEventPolicies interface {
	Approve(ctx context.Context, id string) error
}

// PolicyService handles policy-related API operations.
type PolicyService struct {
	client Client
	scope  log.Scope
}

// Ensure PolicyService implements LogEventPolicies.
var _ LogEventPolicies = (*PolicyService)(nil)

// NewPolicyService creates a new policy service.
func NewPolicyService(client Client, scope log.Scope) *PolicyService {
	return &PolicyService{
		client: client,
		scope:  scope.Child("policies"),
	}
}

// Approve approves a log event policy.
func (s *PolicyService) Approve(ctx context.Context, id string) error {
	s.scope.Debug("approving policy via API", "id", id)

	_, err := s.client.ApproveLogEventPolicy(ctx, id)
	if err != nil {
		s.scope.Error("failed to approve policy", "error", err, "id", id)
		if classified := classifyError(err); classified != nil {
			return fmt.Errorf("approve policy %s: %w", id, classified)
		}
		return err
	}

	s.scope.Debug("approved policy via API", "id", id)
	return nil
}

// Dismiss dismisses a log event policy.
func (s *PolicyService) Dismiss(ctx context.Context, id string) error {
	s.scope.Debug("dismissing policy via API", "id", id)

	_, err := s.client.DismissLogEventPolicy(ctx, id)
	if err != nil {
		s.scope.Error("failed to dismiss policy", "error", err, "id", id)
		if classified := classifyError(err); classified != nil {
			return fmt.Errorf("dismiss policy %s: %w", id, classified)
		}
		return err
	}

	s.scope.Debug("dismissed policy via API", "id", id)
	return nil
}
