package tools

type PolicyApproveInput struct {
	PolicyID string `json:"policy_id"`
}

type PolicyApproveResult struct {
	Approved bool `json:"approved"`
}

func (r PolicyApproveResult) ToMap() map[string]any {
	return map[string]any{"approved": r.Approved}
}
