package block

// EndJourneyInput ends the current journey.
type EndJourneyInput struct{}

// Validate checks that required fields are present.
func (i EndJourneyInput) Validate() error {
	return nil
}
