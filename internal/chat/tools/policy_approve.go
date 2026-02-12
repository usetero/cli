package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/chat"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

// PolicyApproveTool approves log event policies.
type PolicyApproveTool struct {
	policies  sqlite.LogEventPolicies
	getUserID func() string
}

// NewPolicyApproveTool creates a new policy approve tool.
func NewPolicyApproveTool(policies sqlite.LogEventPolicies, getUserID func() string) *PolicyApproveTool {
	return &PolicyApproveTool{policies: policies, getUserID: getUserID}
}

// Name returns the tool name.
func (t *PolicyApproveTool) Name() string {
	return "approve_policy"
}

// Definition returns the tool definition for the chat API.
func (t *PolicyApproveTool) Definition() chat.Tool {
	return chat.Tool{
		Name: t.Name(),
		Description: `Approve a log event policy for enforcement.

When approved, the policy will create exclusion filters in Datadog to reduce log volume.

Use this when the user wants to:
- Approve a specific policy by ID
- Enable a policy recommendation
- Accept a suggested optimization`,
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"policy_id": {
					Type:        "string",
					Description: "The ID of the policy to approve",
				},
			},
			[]string{"policy_id"},
		),
	}
}

// Execute approves the policy and returns the result.
func (t *PolicyApproveTool) Execute(input json.RawMessage) (domaintools.PolicyApproveResult, error) {
	var in domaintools.PolicyApproveInput
	if err := json.Unmarshal(input, &in); err != nil {
		return domaintools.PolicyApproveResult{}, err
	}

	userID := t.getUserID()
	if userID == "" {
		return domaintools.PolicyApproveResult{}, fmt.Errorf("user not authenticated")
	}

	ctx := context.Background()

	err := t.policies.Approve(ctx, in.PolicyID, userID)
	if err != nil {
		return domaintools.PolicyApproveResult{Approved: false}, err
	}

	return domaintools.PolicyApproveResult{Approved: true}, nil
}
