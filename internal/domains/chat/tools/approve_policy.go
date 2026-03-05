package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type ApprovePolicyInput struct {
	PolicyID string `json:"policy_id"`
}

type ApprovePolicyResult struct {
	PolicyID PolicyID `json:"policy_id"`
	Status   string   `json:"status"`
}

type ApprovePolicyTool struct {
	approveFunc func(ctx context.Context, policyID PolicyID) error
}

func NewApprovePolicyTool(approveFunc func(ctx context.Context, policyID PolicyID) error) *ApprovePolicyTool {
	return &ApprovePolicyTool{approveFunc: approveFunc}
}

func (t *ApprovePolicyTool) Definition() Definition {
	return Definition{
		Name:        ApprovePolicyToolName,
		Description: "Approve a pending policy.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"policy_id": map[string]any{"type": "string"},
			},
			"required": []string{"policy_id"},
		},
	}
}

func (t *ApprovePolicyTool) Run(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if t == nil {
		return nil, fmt.Errorf("approve policy tool is not initialized")
	}
	if t.approveFunc == nil {
		return nil, fmt.Errorf("approve policy function is not configured")
	}

	var in ApprovePolicyInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("parse approve policy input: %w", err)
	}
	policyID, err := ParsePolicyID(in.PolicyID)
	if err != nil {
		return nil, err
	}
	if err := t.approveFunc(ctx, policyID); err != nil {
		return nil, err
	}

	return json.Marshal(ApprovePolicyResult{
		PolicyID: policyID,
		Status:   "approved",
	})
}
