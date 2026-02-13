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

// PIIField identifies an attribute path that contains PII and its type.
type PIIField struct {
	Path    []string `json:"path"`
	PIIType string   `json:"pii_type"`
}

// PIILeakageAnalysis is the category-specific analysis for PII leakage policies.
// Stored as JSON in the log_event_policies.analysis column under the "pii_leakage" key.
type PIILeakageAnalysis struct {
	Rationale string     `json:"rationale"`
	RiskLevel string     `json:"risk_level"`
	Benefits  []string   `json:"benefits"`
	Fields    []PIIField `json:"fields"`
}

// PIIAnalysisEnvelope wraps the category-keyed analysis JSON from the database.
type PIIAnalysisEnvelope struct {
	PIILeakage *PIILeakageAnalysis `json:"pii_leakage,omitempty"`
}

// PIIPolicy is a single PII leakage finding with context from joined tables.
type PIIPolicy struct {
	LogEventName string
	ServiceName  string
	RiskLevel    RiskLevel
	Status       PolicyLogStatus
	Fields       []PIIField // Parsed from analysis JSON
}
