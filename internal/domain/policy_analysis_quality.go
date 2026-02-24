package domain

import (
	"fmt"
	"strings"
)

// Quality category slug constants matching control plane schema.
const (
	CategoryDuplicateFields      PolicyCategory = "duplicate_fields"
	CategoryInstrumentationBloat PolicyCategory = "instrumentation_bloat"
	CategoryOversizedFields      PolicyCategory = "oversized_fields"
	CategoryWrongLevel           PolicyCategory = "wrong_level"
)

// DuplicateFieldsAnalysis identifies fields with the same data stored
// redundantly under different names within a single log event.
type DuplicateFieldsAnalysis struct {
	baseAnalysis
	Pairs []DuplicatePair `json:"pairs"` // Duplicate field pairs to deduplicate
}

func (a DuplicateFieldsAnalysis) Category() PolicyCategory { return CategoryDuplicateFields }

func (a DuplicateFieldsAnalysis) Subtitle() string {
	if len(a.Pairs) > 0 {
		return fmt.Sprintf("%d duplicate field pairs", len(a.Pairs))
	}
	return ""
}

func (a DuplicateFieldsAnalysis) ActionDetail() string {
	count := 0
	for _, p := range a.Pairs {
		count += len(p.Remove)
	}
	if count > 0 {
		return fmt.Sprintf("%d duplicate fields", count)
	}
	return ""
}

func (a DuplicateFieldsAnalysis) RelevantKeys() []FieldPath {
	var keys []FieldPath
	for _, p := range a.Pairs {
		keys = append(keys, p.Remove...)
		if !p.Keep.IsEmpty() {
			keys = append(keys, p.Keep)
		}
	}
	return keys
}

// InstrumentationBloatAnalysis identifies SDK/collector metadata fields
// that no engineer added intentionally and no engineer would ever query.
type InstrumentationBloatAnalysis struct {
	baseAnalysis
	Fields []FieldPath `json:"fields"` // Attribute paths to remove
}

func (a InstrumentationBloatAnalysis) Category() PolicyCategory { return CategoryInstrumentationBloat }

func (a InstrumentationBloatAnalysis) Subtitle() string {
	if len(a.Fields) > 0 {
		return fmt.Sprintf("%d fields to remove", len(a.Fields))
	}
	return ""
}

func (a InstrumentationBloatAnalysis) ActionDetail() string {
	if len(a.Fields) > 0 {
		return fmt.Sprintf("%d bloat fields", len(a.Fields))
	}
	return ""
}

func (a InstrumentationBloatAnalysis) RelevantKeys() []FieldPath {
	return a.Fields
}

// OversizedFieldsAnalysis identifies fields where the value is disproportionately
// large relative to its diagnostic utility.
type OversizedFieldsAnalysis struct {
	baseAnalysis
	Fields []FieldPath `json:"fields"` // Attribute paths to truncate
}

func (a OversizedFieldsAnalysis) Category() PolicyCategory { return CategoryOversizedFields }

func (a OversizedFieldsAnalysis) Subtitle() string {
	if len(a.Fields) > 0 {
		return fmt.Sprintf("%d fields to truncate", len(a.Fields))
	}
	return ""
}

func (a OversizedFieldsAnalysis) ActionDetail() string {
	if len(a.Fields) > 0 {
		return fmt.Sprintf("%d oversized fields", len(a.Fields))
	}
	return ""
}

func (a OversizedFieldsAnalysis) RelevantKeys() []FieldPath {
	return a.Fields
}

// WrongLevelAnalysis identifies events emitted at the wrong severity level.
type WrongLevelAnalysis struct {
	baseAnalysis
	CurrentLevel   string `json:"current_level"`   // Normalized level the event currently uses (debug, info, warn, error)
	SuggestedLevel string `json:"suggested_level"` // Normalized level the event should use
}

func (a WrongLevelAnalysis) Category() PolicyCategory { return CategoryWrongLevel }

func (a WrongLevelAnalysis) Subtitle() string {
	if a.CurrentLevel != "" && a.SuggestedLevel != "" {
		return fmt.Sprintf("%s → %s", strings.ToUpper(a.CurrentLevel), strings.ToUpper(a.SuggestedLevel))
	}
	return ""
}

func (a WrongLevelAnalysis) ActionDetail() string {
	if a.SuggestedLevel != "" {
		return fmt.Sprintf("re-level to %s", a.SuggestedLevel)
	}
	return ""
}

func (a WrongLevelAnalysis) RelevantKeys() []FieldPath { return nil }
