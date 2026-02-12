package tools

// StartPolicyApprovalResult holds the result of starting a policy approval wizard.
type StartPolicyApprovalResult struct {
	Started bool
}

// ToMap serializes the result for the GraphQL API.
func (r StartPolicyApprovalResult) ToMap() map[string]any {
	return map[string]any{"started": r.Started}
}
