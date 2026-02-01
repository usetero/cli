package block

import "fmt"

// AddContextInput adds an entity to the conversation context.
type AddContextInput struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

// Validate checks that required fields are present.
func (i AddContextInput) Validate() error {
	if i.EntityType == "" {
		return fmt.Errorf("entity_type is required")
	}
	if i.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	return nil
}

// AddContextResult is the result of adding context.
type AddContextResult struct {
	Added bool `json:"added"`
}
