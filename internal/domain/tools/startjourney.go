package tools

// StartJourneyInput is the input schema for the start_journey tool.
type StartJourneyInput struct {
	Name string `json:"name"`
}

// StartJourneyResult is the typed output of a start_journey tool execution.
type StartJourneyResult struct {
	Started bool `json:"started"`
}

// ToMap serializes the result for the GraphQL API.
func (r StartJourneyResult) ToMap() map[string]any {
	return map[string]any{"started": r.Started}
}
