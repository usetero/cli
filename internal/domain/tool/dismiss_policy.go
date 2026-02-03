package tool

import "github.com/google/uuid"

type DismissPolicyInput struct {
	PolicyID uuid.UUID `json:"policy_id"`
}

type DismissPolicyResult struct {
	PolicyID  uuid.UUID `json:"policy_id"`
	Dismissed bool      `json:"dismissed"`
}
