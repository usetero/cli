package domain

// PolicyCard contains all data needed to render a rich policy card.
// Fetched from log_event_policy_statuses_cache with enrichment JOINs.
type PolicyCard struct {
	PolicyID               string
	ServiceName            string
	LogEventName           string
	Category               string
	CategoryType           string
	Action                 string
	Status                 string
	Severity               string
	CategoryDisplayName    string
	VolumePerHour          *float64
	BytesPerHour           *float64
	EstimatedCostPerHour   *float64
	EstimatedVolumePerHour *float64
	EstimatedBytesPerHour  *float64
	SurvivalRate           *float64
	Analysis               string // Raw category-keyed JSON from log_event_policies.analysis
	Examples               string // Raw JSON array from log_events.examples

	// Event-level baselines from log_events.
	EventBaselineAvgBytes      *float64 // Trailing 7-day avg bytes per event
	EventBaselineVolumePerHour *float64 // Trailing 7-day volume per hour

	// Per-field byte sizes from log_event_fields, keyed by dot-path (e.g. "http.status").
	// Populated by the sqlite layer for quality categories. nil for other categories.
	FieldSizes map[string]float64
}
