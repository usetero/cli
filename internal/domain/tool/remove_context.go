package tool

import "github.com/google/uuid"

type RemoveContextInput struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
}

type RemoveContextResult struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	Removed    bool      `json:"removed"`
}
