package tools

// EndJourneyInput is the input schema for the end_journey tool.
type EndJourneyInput struct{}

// EndJourneyResult is the typed output of an end_journey tool execution.
type EndJourneyResult struct {
	Ended bool `json:"ended"`
}

// ToMap serializes the result for the GraphQL API.
func (r EndJourneyResult) ToMap() map[string]any {
	return map[string]any{"ended": r.Ended}
}
