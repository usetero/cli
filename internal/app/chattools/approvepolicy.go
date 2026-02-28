package tools

import (
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/api/chatclient"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

// NewApprovePolicyAction creates an ActionTool for approve_policy.
// userIDFunc is called at execution time to get the current user's ID.
func NewApprovePolicyAction(db sqlite.DB, userIDFunc func() string) ActionTool {
	def := chat.Tool{
		Name:        "approve_policy",
		Description: "Approve a log event policy for enforcement. This marks the policy as approved and queues it for the enforcement pipeline.",
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"policy_id": {
					Type:        "string",
					Description: "UUID of the policy (e.g., '4a3b1c2d-...'). Use the query tool to look up policy IDs: SELECT policy_id, service_name, log_event_name, category FROM log_event_policy_statuses_cache WHERE status = 'PENDING'",
				},
			},
			[]string{"policy_id"},
		),
	}

	executor := func(input json.RawMessage) (tools.Result, error) {
		var in tools.ApprovePolicyInput
		if err := json.Unmarshal(input, &in); err != nil {
			return tools.Result{}, err
		}

		ctx, cancel := withToolTimeout()
		defer cancel()

		if err := db.LogEventPolicies().Approve(ctx, in.PolicyID, userIDFunc()); err != nil {
			return tools.Result{}, fmt.Errorf("approve policy: %w", err)
		}

		return tools.Result{
			Content: tools.ApprovePolicyResult{
				PolicyID: in.PolicyID,
				Status:   "APPROVED",
			}.ToMap(),
		}, nil
	}

	config := action.Config{
		DisplayName: func(_ json.RawMessage) string {
			return "Approve Policy"
		},
		Status: func(input json.RawMessage) string {
			var in tools.ApprovePolicyInput
			if json.Unmarshal(input, &in) != nil {
				return ""
			}
			return fmt.Sprintf("Approving %s", in.PolicyID)
		},
		Result: func(result tools.Result) string {
			id, _ := result.Content["policy_id"].(string)
			return fmt.Sprintf("Policy %s approved", id)
		},
	}

	return NewActionTool(def, executor, config)
}
