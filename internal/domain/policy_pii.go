package domain

// PIIType constants for categorising PII fields.
const (
	PIITypeEmail       = "email"
	PIITypeName        = "name"
	PIITypePhone       = "phone"
	PIITypeAddress     = "address"
	PIITypeCreditCard  = "credit_card"
	PIITypeSSN         = "ssn"
	PIITypePassword    = "password"
	PIITypeIPAddress   = "ip_address"
	PIITypeDateOfBirth = "date_of_birth"
	PIITypeGeneral     = "general"
)

// PIIField identifies an attribute path that contains PII and its types.
type PIIField struct {
	Path     []string `json:"path"`
	PIITypes []string `json:"pii_types"`
	Observed bool     `json:"observed"` // True if actual PII was seen in log values; false means at-risk based on name/context.
}

// PIILeakageAnalysis is the category-specific analysis for PII leakage policies.
// Stored as JSON in the log_event_policies.analysis column under the "pii_leakage" key.
type PIILeakageAnalysis struct {
	Rationale string     `json:"rationale"`
	Benefits  []string   `json:"benefits"`
	Fields    []PIIField `json:"fields"`
}

// PIIAnalysisEnvelope wraps the category-keyed analysis JSON from the database.
type PIIAnalysisEnvelope struct {
	PIILeakage *PIILeakageAnalysis `json:"pii_leakage,omitempty"`
}

// PIIPolicy is a single pending PII leakage finding with context from joined tables.
type PIIPolicy struct {
	LogEventName  string
	ServiceName   string
	Fields        []PIIField // Parsed from analysis JSON
	VolumePerHour float64    // Log event volume from log_event_statuses_cache
	AnyObserved   bool       // True if any field has observed PII, computed in SQL
	HasVolumes    bool       // Whether volume data exists for this log event
}
