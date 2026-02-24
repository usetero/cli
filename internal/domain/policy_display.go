package domain

import (
	"fmt"
	"math"
	"strings"

	"github.com/usetero/cli/internal/format"
)

// PolicyImpact holds the before/after metrics for a policy's estimated impact.
type PolicyImpact struct {
	VolumeFrom  string // "883.7k evt/hr"
	VolumeTo    string // "~8.8k evt/hr"
	VolumePct   string // "▼ 99%"
	StorageFrom string // "3.5 GB/hr"
	StorageTo   string // "~35 MB/hr"
	StoragePct  string // "▼ 99%"
	Savings     string // "~$16.3k/yr"
}

// Headline returns the one-line action description for this policy.
// Examples: "Sample — keep ~1% of volume", "Drop this event", "Trim — 3 bloat fields".
func (p *Policy) Headline() string {
	var subtitle, detail string
	if p.Analysis != nil {
		subtitle = p.Analysis.Subtitle()
		detail = p.Analysis.ActionDetail()
	}

	switch p.Action {
	case PolicyActionDrop:
		return "Drop this event"
	case PolicyActionSample:
		if subtitle != "" && strings.HasPrefix(subtitle, "Sample") {
			return subtitle
		}
		if p.SurvivalRate != nil && *p.SurvivalRate > 0 {
			pct := *p.SurvivalRate * 100
			if pct < 1 {
				return "Sample — keep <1% of volume"
			}
			return fmt.Sprintf("Sample — keep ~%.0f%% of volume", pct)
		}
		return "Sample at a reduced rate"
	case PolicyActionFilter:
		if detail != "" {
			return "Filter — " + detail
		}
		return "Filter a subset"
	case PolicyActionTrim:
		if detail != "" {
			return "Trim — " + detail
		}
		return "Remove fields"
	case PolicyActionNone:
		if p.Category == CategoryWrongLevel && subtitle != "" {
			return "Re-level " + subtitle
		}
		return "Informational"
	default:
		return ""
	}
}

// Mechanism explains what approving this policy does.
// Examples: "Approving keeps 1 in every ~8,837 events.", "Approving creates a pipeline rule that drops this event before ingestion."
func (p *Policy) Mechanism() string {
	switch p.Action {
	case PolicyActionDrop:
		return "Approving creates a pipeline rule that drops this event before ingestion."
	case PolicyActionSample:
		if p.SurvivalRate != nil && *p.SurvivalRate > 0 {
			n := int(math.Round(1.0 / *p.SurvivalRate))
			return fmt.Sprintf("Approving keeps 1 in every ~%s events.", format.Volume(float64(n)))
		}
		return "Approving creates a pipeline rule that samples this event."
	case PolicyActionFilter:
		return "Approving creates a pipeline rule that filters matching events."
	case PolicyActionTrim:
		return "Approving creates a pipeline rule that removes these fields before ingestion."
	case PolicyActionNone:
		return "This is an informational finding. No pipeline rule will be created."
	default:
		return ""
	}
}

// Pitch returns the first sentence of the analysis rationale.
func (p *Policy) Pitch() string {
	if p.Analysis == nil {
		return ""
	}
	return format.FirstSentence(p.Analysis.Rationale())
}

// CostPerYear returns the formatted estimated annual cost savings, or "" if not applicable.
// Only meaningful for waste and quality categories.
func (p *Policy) CostPerYear() string {
	if p.CategoryType == CategoryTypeCompliance {
		return ""
	}
	if p.EstimatedCostPerHour != nil && *p.EstimatedCostPerHour > 0 {
		return "~" + format.YearlyCost(*p.EstimatedCostPerHour)
	}
	return ""
}

// Impact returns the before/after metrics for this policy, or nil if no estimates exist.
func (p *Policy) Impact() *PolicyImpact {
	hasCost := p.EstimatedCostPerHour != nil && *p.EstimatedCostPerHour > 0
	hasVolume := p.VolumePerHour != nil && p.EstimatedVolumePerHour != nil
	hasBytes := p.BytesPerHour != nil && p.EstimatedBytesPerHour != nil

	if !hasCost && !hasVolume && !hasBytes {
		return nil
	}

	impact := &PolicyImpact{}

	if hasVolume {
		impact.VolumeFrom = format.Volume(*p.VolumePerHour) + " evt/hr"
		impact.VolumeTo = "~" + format.Volume(*p.VolumePerHour-*p.EstimatedVolumePerHour) + " evt/hr"
		impact.VolumePct = reductionPct(*p.EstimatedVolumePerHour, *p.VolumePerHour)
	}

	if hasBytes {
		impact.StorageFrom = format.Bytes(*p.BytesPerHour) + "/hr"
		impact.StorageTo = "~" + format.Bytes(*p.BytesPerHour-*p.EstimatedBytesPerHour) + "/hr"
		impact.StoragePct = reductionPct(*p.EstimatedBytesPerHour, *p.BytesPerHour)
	}

	if hasCost {
		impact.Savings = "~" + format.YearlyCost(*p.EstimatedCostPerHour)
	}

	return impact
}

func reductionPct(reduction, total float64) string {
	if total == 0 {
		return ""
	}
	pct := (reduction / total) * 100
	if pct >= 99.5 {
		return "▼ 99%+"
	}
	return fmt.Sprintf("▼ %.0f%%", pct)
}
