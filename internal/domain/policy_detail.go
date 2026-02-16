package domain

// PolicyDetail represents an individual pending policy within a category.
type PolicyDetail struct {
	PolicyID               string
	LogEventName           string
	Benefits               string
	EstimatedCostPerHour   float64
	EstimatedVolumePerHour float64
}
