package tool

import "github.com/google/uuid"

type ApprovePolicyInput struct {
	PolicyID uuid.UUID `json:"policy_id"`
}

type ApprovePolicyResult struct {
	PolicyID uuid.UUID `json:"policy_id"`
	Approved bool      `json:"approved"`
}
