package block

import "fmt"

// StartJourneyInput starts a guided journey.
type StartJourneyInput struct {
	Name string `json:"name"`
}

// Validate checks that required fields are present.
func (i StartJourneyInput) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
