package block

import "fmt"

// DismissPolicyInput dismisses a policy recommendation.
type DismissPolicyInput struct {
	PolicyID string `json:"policy_id"`
	Reason   string `json:"reason,omitempty"`
}

// Validate checks that required fields are present.
func (i DismissPolicyInput) Validate() error {
	if i.PolicyID == "" {
		return fmt.Errorf("policy_id is required")
	}
	return nil
}

// DismissPolicyResult is the result of dismissing a policy.
type DismissPolicyResult struct {
	Dismissed bool `json:"dismissed"`
}
