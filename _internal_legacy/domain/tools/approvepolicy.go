package tools

// ApprovePolicyInput is the input schema for the approve_policy tool.
type ApprovePolicyInput struct {
	PolicyID string `json:"policy_id"`
}

// ApprovePolicyResult is the typed output of an approve_policy tool execution.
type ApprovePolicyResult struct {
	PolicyID string `json:"policy_id"`
	Status   string `json:"status"`
}

// ToMap serializes the result for the GraphQL API.
func (r ApprovePolicyResult) ToMap() map[string]any {
	return map[string]any{
		"policy_id": r.PolicyID,
		"status":    r.Status,
	}
}
