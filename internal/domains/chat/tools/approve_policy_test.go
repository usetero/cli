package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestApprovePolicyTool_Run(t *testing.T) {
	t.Parallel()

	approved := PolicyID("")
	tool := NewApprovePolicyTool(func(_ context.Context, policyID PolicyID) error {
		approved = policyID
		return nil
	})

	out, err := tool.Run(context.Background(), json.RawMessage(`{"policy_id":"pol_1"}`))
	if err != nil {
		t.Fatalf("run approve policy: %v", err)
	}
	if approved != "pol_1" {
		t.Fatalf("unexpected approved policy id: %q", approved)
	}

	var parsed ApprovePolicyResult
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if parsed.PolicyID != "pol_1" || parsed.Status != "approved" {
		t.Fatalf("unexpected output: %+v", parsed)
	}
}

func TestApprovePolicyTool_RejectsMissingPolicyID(t *testing.T) {
	t.Parallel()

	tool := NewApprovePolicyTool(func(context.Context, PolicyID) error { return nil })
	_, err := tool.Run(context.Background(), json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "policy_id is required") {
		t.Fatalf("expected missing policy_id error, got %v", err)
	}
}
