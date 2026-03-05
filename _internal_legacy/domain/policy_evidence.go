package domain

import (
	"maps"
	"slices"
)

// Evidence is a marker interface for polymorphic evidence types.
// Each evidence shape represents a different way of presenting log
// example data to prove a policy's findings. Consumers type-switch
// to render the appropriate shape.
type Evidence interface{ evidenceKind() string }

// ConstantVariesEvidence shows which fields are identical across
// multiple log examples and which vary. Proves that instances are
// near-identical. Used by broken_records and commodity_traffic.
type ConstantVariesEvidence struct {
	Constant     []FieldValue
	Varying      []VaryingField
	ExampleCount int
}

func (ConstantVariesEvidence) evidenceKind() string { return "constant_varies" }

// FieldValue is a single key-value pair from a flattened log attribute.
type FieldValue struct {
	Key   string
	Value string
}

// VaryingField is a key with multiple distinct values across examples.
type VaryingField struct {
	Key    string
	Values []string // deduplicated, up to 4
}

// HighlightedExampleEvidence shows a single log example with specific
// fields called out as relevant to the finding. Used by compliance
// (sensitive fields), bot_traffic (user-agent field), and as a fallback
// for field-targeting quality categories when byte sizes aren't available.
type HighlightedExampleEvidence struct {
	Attrs        []FieldValue // All attributes from the example, ordered with relevant first
	RelevantKeys []FieldPath  // Keys that should be visually highlighted
}

func (HighlightedExampleEvidence) evidenceKind() string { return "highlighted_example" }

// FieldListEvidence lists the fields a quality policy targets, each with
// its measured byte size. Used by instrumentation_bloat, oversized_fields,
// and duplicate_fields when per-field byte data is available.
// Fields are sorted by BytesPerEvent descending (biggest impact first).
type FieldListEvidence struct {
	Fields        []FieldSize
	TotalBytes    float64 // Sum of targeted fields' bytes
	EventAvgBytes float64 // Whole-event average bytes (baseline)
	BytesFraction float64 // TotalBytes / EventAvgBytes
}

func (FieldListEvidence) evidenceKind() string { return "field_list" }

// FieldSize is a single field targeted by a quality policy with its measured byte impact.
type FieldSize struct {
	Key           string  // Dot-key (e.g. "http.request.body")
	BytesPerEvent float64 // Average bytes this field contributes per event
}

// BuildEvidence derives the evidence structure for a policy from its
// analysis type and log examples. Returns nil if no evidence can be
// built (no examples, unsupported category, etc.).
func BuildEvidence(p *Policy) Evidence {
	if p.Analysis == nil {
		return nil
	}

	switch p.Analysis.(type) {
	case BrokenRecordsAnalysis, CommodityTrafficAnalysis:
		if len(p.Examples) >= 2 {
			return buildConstantVaries(p.Examples)
		}
		return nil

	case DuplicateFieldsAnalysis, InstrumentationBloatAnalysis, OversizedFieldsAnalysis:
		if ev := buildFieldList(p); ev != nil {
			return ev
		}
		return buildHighlightedExample(p)

	case BotTrafficAnalysis,
		PIILeakageAnalysis, SecretsLeakageAnalysis, PHILeakageAnalysis, PaymentDataLeakageAnalysis:
		return buildHighlightedExample(p)

	default:
		// health_checks, debug_artifacts, malformed, redundant_events,
		// dead_weight, wrong_level — rationale-only, no evidence.
		return nil
	}
}

// buildConstantVaries compares fields across multiple log examples,
// splitting them into constant (same value in every example) and
// varying (different values across examples).
func buildConstantVaries(examples []LogExample) Evidence {
	type flatEx struct{ attrs map[string]string }
	flats := make([]flatEx, len(examples))
	allKeys := make(map[string]bool)

	for i, ex := range examples {
		flat := make(map[string]string)
		for _, attr := range ex.FlatAttrs(nil) {
			flat[attr.Key] = attr.Value
			allKeys[attr.Key] = true
		}
		flats[i] = flatEx{attrs: flat}
	}

	var constant []FieldValue
	var varying []VaryingField

	for _, key := range slices.Sorted(maps.Keys(allKeys)) {
		firstVal := flats[0].attrs[key]
		isConstant := true
		values := []string{firstVal}

		for _, f := range flats[1:] {
			v := f.attrs[key]
			values = append(values, v)
			if v != firstVal {
				isConstant = false
			}
		}

		if isConstant && firstVal != "" {
			constant = append(constant, FieldValue{Key: key, Value: firstVal})
		} else {
			seen := make(map[string]bool)
			var unique []string
			for _, s := range values {
				if s != "" && !seen[s] {
					seen[s] = true
					unique = append(unique, s)
					if len(unique) >= 4 {
						break
					}
				}
			}
			varying = append(varying, VaryingField{Key: key, Values: unique})
		}
	}

	if len(constant) == 0 && len(varying) == 0 {
		return nil
	}
	return &ConstantVariesEvidence{
		Constant:     constant,
		Varying:      varying,
		ExampleCount: len(examples),
	}
}

// buildFieldList builds a FieldListEvidence from the policy's relevant keys
// and field sizes. Returns nil if field sizes aren't populated.
func buildFieldList(p *Policy) Evidence {
	if len(p.FieldSizes) == 0 || len(p.RelevantKeys) == 0 {
		return nil
	}

	var fields []FieldSize
	var totalBytes float64
	for _, rk := range p.RelevantKeys {
		key := rk.Key()
		if bytes, ok := p.FieldSizes[key]; ok {
			fields = append(fields, FieldSize{Key: key, BytesPerEvent: bytes})
			totalBytes += bytes
		}
	}
	if len(fields) == 0 {
		return nil
	}

	// Sort by bytes descending (biggest impact first).
	slices.SortFunc(fields, func(a, b FieldSize) int {
		if a.BytesPerEvent > b.BytesPerEvent {
			return -1
		}
		if a.BytesPerEvent < b.BytesPerEvent {
			return 1
		}
		return 0
	})

	ev := &FieldListEvidence{
		Fields:     fields,
		TotalBytes: totalBytes,
	}
	if p.EventBaselineAvgBytes != nil && *p.EventBaselineAvgBytes > 0 {
		ev.EventAvgBytes = *p.EventBaselineAvgBytes
		ev.BytesFraction = totalBytes / *p.EventBaselineAvgBytes
	}
	return ev
}

// buildHighlightedExample takes the first log example and returns its
// attributes ordered with relevant keys first, plus the list of keys
// that should be visually highlighted.
func buildHighlightedExample(p *Policy) Evidence {
	if len(p.Examples) == 0 {
		return nil
	}
	relevantKeys := p.RelevantKeys
	attrs := p.Examples[0].FlatAttrs(relevantKeys)
	if len(attrs) == 0 {
		return nil
	}
	fvs := make([]FieldValue, len(attrs))
	for i, a := range attrs {
		fvs[i] = FieldValue(a)
	}
	return &HighlightedExampleEvidence{
		Attrs:        fvs,
		RelevantKeys: relevantKeys,
	}
}
