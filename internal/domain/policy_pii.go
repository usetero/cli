package domain

// PIILeakageAnalysis is the category-specific analysis for PII leakage policies.
// Stored as JSON in the log_event_policies.analysis column under the "pii_leakage" key.
type PIILeakageAnalysis struct {
	Rationale string     `json:"rationale"`
	RiskLevel string     `json:"risk_level"`
	Benefits  []string   `json:"benefits"`
	Fields    [][]string `json:"fields"` // Attribute paths containing PII, e.g. [["attributes","user","email"]]
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
	Fields       [][]string // Parsed from analysis JSON
}
