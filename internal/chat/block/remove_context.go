package block

import "fmt"

// RemoveContextInput removes an entity from the conversation context.
type RemoveContextInput struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

// Validate checks that required fields are present.
func (i RemoveContextInput) Validate() error {
	if i.EntityType == "" {
		return fmt.Errorf("entity_type is required")
	}
	if i.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	return nil
}

// RemoveContextResult is the result of removing context.
type RemoveContextResult struct {
	Removed bool `json:"removed"`
}
