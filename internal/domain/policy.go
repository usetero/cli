package domain

// PolicyID is a unique identifier for a policy.
type PolicyID string

func (id PolicyID) String() string { return string(id) }

// Policy is a fully parsed policy ready for any consumer.
// Built from PolicyCard (the raw DB type) via ParsePolicy.
// Analysis is typed per category, examples are parsed, and relevant keys are extracted.
type Policy struct {
	ID                  PolicyID
	ServiceName         string
	LogEventName        string
	Category            PolicyCategory
	CategoryType        CategoryType
	Action              PolicyAction
	Status              PolicyStatus
	Severity            PolicySeverity
	CategoryDisplayName string

	// Metrics (nil = not measured)
	VolumePerHour          *float64
	BytesPerHour           *float64
	EstimatedCostPerHour   *float64
	EstimatedVolumePerHour *float64
	EstimatedBytesPerHour  *float64
	SurvivalRate           *float64

	// Event-level baselines from log_events.
	EventBaselineAvgBytes      *float64 // Trailing 7-day avg bytes per event
	EventBaselineVolumePerHour *float64 // Trailing 7-day volume per hour

	// Per-field byte sizes from log_event_fields, keyed by dot-path (e.g. "http.status").
	// Populated for quality categories via PolicyCard. nil for other categories.
	FieldSizes map[string]float64

	// Parsed from JSON
	Analysis     PolicyAnalysis // nil if analysis JSON is empty or unparseable
	Examples     []LogExample   // parsed, infrastructure stripped
	RelevantKeys []FieldPath    // attribute paths relevant to this category
}

// ParsePolicy converts a raw PolicyCard (from the DB) into a fully parsed Policy.
func ParsePolicy(card *PolicyCard) *Policy {
	p := &Policy{
		ID:                         PolicyID(card.PolicyID),
		ServiceName:                card.ServiceName,
		LogEventName:               card.LogEventName,
		Category:                   PolicyCategory(card.Category),
		CategoryType:               CategoryType(card.CategoryType),
		Action:                     PolicyAction(card.Action),
		Status:                     PolicyStatus(card.Status),
		Severity:                   PolicySeverity(card.Severity),
		CategoryDisplayName:        card.CategoryDisplayName,
		VolumePerHour:              card.VolumePerHour,
		BytesPerHour:               card.BytesPerHour,
		EstimatedCostPerHour:       card.EstimatedCostPerHour,
		EstimatedVolumePerHour:     card.EstimatedVolumePerHour,
		EstimatedBytesPerHour:      card.EstimatedBytesPerHour,
		SurvivalRate:               card.SurvivalRate,
		EventBaselineAvgBytes:      card.EventBaselineAvgBytes,
		EventBaselineVolumePerHour: card.EventBaselineVolumePerHour,
		FieldSizes:                 card.FieldSizes,
	}

	p.Analysis = parseAnalysis(card.Analysis, PolicyCategory(card.Category))
	if p.Analysis != nil {
		p.RelevantKeys = p.Analysis.RelevantKeys()
	}

	if card.Examples != "" {
		p.Examples = ParseLogExamples(card.Examples)
	}

	return p
}
