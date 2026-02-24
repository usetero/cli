package domain

import (
	"encoding/json"
	"fmt"
)

// PolicySeverity is the compliance severity level of a policy.
// Computed by the control plane and stored on the policy row.
type PolicySeverity string

const (
	SeverityLow      PolicySeverity = "low"
	SeverityMedium   PolicySeverity = "medium"
	SeverityHigh     PolicySeverity = "high"
	SeverityCritical PolicySeverity = "critical"
)

// DuplicatePair identifies a set of duplicate fields — one to keep, others to remove.
type DuplicatePair struct {
	Remove []FieldPath `json:"remove"` // Duplicate field paths to remove
	Keep   FieldPath   `json:"keep"`   // Canonical field path to keep
}

// AnalysisEnvelope wraps the category-keyed analysis JSON from log_event_policies.analysis.
// The column stores {"category_name": { ...analysis... }} — one key per policy.
type AnalysisEnvelope map[string]json.RawMessage

// Parse extracts and unmarshals the analysis for a given category into the target type.
func (e AnalysisEnvelope) Parse(category string, target any) error {
	raw, ok := e[category]
	if !ok {
		return fmt.Errorf("no analysis for category %q", category)
	}
	return json.Unmarshal(raw, target)
}

// PolicyAnalysis is the parsed, typed analysis for a policy.
// Each category implements this interface on its analysis struct.
type PolicyAnalysis interface {
	Category() PolicyCategory
	Rationale() string
	Subtitle() string
	ActionDetail() string
	RelevantKeys() []FieldPath
}

// baseAnalysis provides default implementations for rationale-only categories.
// Embed in analysis types that have no category-specific fields beyond rationale.
type baseAnalysis struct {
	RationaleText string `json:"rationale"`
}

func (a baseAnalysis) Rationale() string         { return a.RationaleText }
func (a baseAnalysis) Subtitle() string          { return "" }
func (a baseAnalysis) ActionDetail() string      { return "" }
func (a baseAnalysis) RelevantKeys() []FieldPath { return nil }

// parseAnalysis unmarshals the analysis JSON for a category into a typed PolicyAnalysis.
// Returns nil if the JSON is empty or the category is unknown.
func parseAnalysis(analysisJSON string, category PolicyCategory) PolicyAnalysis {
	if analysisJSON == "" {
		return nil
	}
	var envelope AnalysisEnvelope
	if err := json.Unmarshal([]byte(analysisJSON), &envelope); err != nil {
		return nil
	}
	raw, ok := envelope[string(category)]
	if !ok {
		return nil
	}

	var target PolicyAnalysis
	switch category {
	case CategoryHealthChecks:
		var a HealthChecksAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryBotTraffic:
		var a BotTrafficAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryDebugArtifacts:
		var a DebugArtifactsAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryMalformed:
		var a MalformedAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryBrokenRecords:
		var a BrokenRecordsAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryCommodityTraffic:
		var a CommodityTrafficAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryRedundantEvents:
		var a RedundantEventsAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryDeadWeight:
		var a DeadWeightAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryDuplicateFields:
		var a DuplicateFieldsAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryInstrumentationBloat:
		var a InstrumentationBloatAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryOversizedFields:
		var a OversizedFieldsAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryWrongLevel:
		var a WrongLevelAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryPIILeakage:
		var a PIILeakageAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategorySecretsLeakage:
		var a SecretsLeakageAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryPHILeakage:
		var a PHILeakageAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	case CategoryPaymentDataLeakage:
		var a PaymentDataLeakageAnalysis
		if json.Unmarshal(raw, &a) == nil {
			target = a
		}
	}
	return target
}

// FormatInterval renders seconds as a human-friendly duration.
func FormatInterval(seconds int) string {
	switch {
	case seconds >= 3600 && seconds%3600 == 0:
		return fmt.Sprintf("%dh", seconds/3600)
	case seconds >= 60 && seconds%60 == 0:
		return fmt.Sprintf("%dm", seconds/60)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}
