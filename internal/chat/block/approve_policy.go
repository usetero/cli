package block

import "fmt"

// ApprovePolicyInput approves a policy recommendation.
type ApprovePolicyInput struct {
	PolicyID string `json:"policy_id"`
}

// Validate checks that required fields are present.
func (i ApprovePolicyInput) Validate() error {
	if i.PolicyID == "" {
		return fmt.Errorf("policy_id is required")
	}
	return nil
}

// ApprovePolicyResult is the result of approving a policy.
type ApprovePolicyResult struct {
	Approved bool `json:"approved"`
}
