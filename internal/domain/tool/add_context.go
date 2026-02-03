package tool

import "github.com/google/uuid"

type AddContextInput struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
}

type AddContextResult struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	Added      bool      `json:"added"`
}
